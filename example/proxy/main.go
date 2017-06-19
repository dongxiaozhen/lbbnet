package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbnet"
)

var cfg lbbconsul.ConsulConfig
var foundServer string
var cproxy = &Cproxy{}
var sproxy = &Sproxy{}
var SM = newServerManager()
var CM = newCManager()

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 2222, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "proxy_id_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "server_proxy", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:2221", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "serverNode_2", "found server name")
	flag.Parse()
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)
	// SM = newServerManager()

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		fmt.Println("open return", err)
		return
	}

	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port), sproxy, 30*time.Second)
	if err != nil {
		fmt.Println("create server err", err)
		return
	}

	go func() {
		tick := time.NewTicker(2 * time.Second)
		var oldSer = make(map[string]*lbbconsul.ServiceInfo)

		for range tick.C {
			err := lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
			if err != nil {
				fmt.Println("discover server err", foundServer)
				continue
			}
			services, ok := lbbconsul.GConsulClient.GetAllService(foundServer)
			if !ok {
				fmt.Println("not find server err", foundServer)
				continue
			}
			// for k, v := range services {
			// if _, ok := oldSer[k]; !ok {
			// fmt.Println("make ", k, *v)
			// go func(s *lbbconsul.ServiceInfo) {
			// _, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", s.IP, s.Port), cproxy, 60*time.Second)
			// if err != nil {
			// fmt.Println("proxy client err", err)
			// }
			// }(v)
			// }
			// }
			compareDiff(oldSer, services)
			oldSer = services
		}
	}()

	<-exist
	s.Close()
	lbbconsul.GConsulClient.Close()
}

func compareDiff(old, new map[string]*lbbconsul.ServiceInfo) {
	for k, v := range old {
		if v2, ok := new[k]; ok {
			if v2.IP == v.IP && v2.Port == v.Port {
			} else {
				fmt.Println("-------------------remvoe server", *v)
				CM.RemoveServerByAddr(fmt.Sprintf("%s:%d", v.IP, v.Port))
				t, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", v2.IP, v2.Port), cproxy, 60*time.Second)
				if err != nil {
					fmt.Println("proxy client err", err)
				} else {
					CM.AddTClient(fmt.Sprintf("%s:%d", v2.IP, v2.Port), t)
				}
			}
		} else {
			fmt.Println("-------------------remove server", *v)
			CM.RemoveServerByAddr(fmt.Sprintf("%s:%d", v.IP, v.Port))
		}
	}

	for k, v := range new {
		if _, ok := old[k]; !ok {
			fmt.Println("-------------------add server", *v)
			t, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), cproxy, 60*time.Second)
			if err != nil {
				fmt.Println("proxy client err", err)
			} else {
				CM.AddTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), t)
			}
		}
	}
}

type ServerManager struct {
	seq     uint32
	clients map[*lbbnet.Transport]uint32
	sync.Mutex
}

func newServerManager() *ServerManager {
	return &ServerManager{
		seq:     0,
		clients: make(map[*lbbnet.Transport]uint32),
	}
}

func (s *ServerManager) RemoveServer(t *lbbnet.Transport) {
	s.Lock()
	defer s.Unlock()
	t.Close()
	delete(s.clients, t)
}

func (s *ServerManager) AddServer(conn *lbbnet.Transport) {
	s.Lock()
	defer s.Unlock()
	s.seq++
	s.clients[conn] = s.seq
}

func (s *ServerManager) GetService(t *lbbnet.Transport) uint32 {
	s.Lock()
	defer s.Unlock()
	return s.clients[t]
}

func (s *ServerManager) GetServiceById(sid uint32) *lbbnet.Transport {
	s.Lock()
	defer s.Unlock()
	for t, id := range s.clients {
		if id == sid {
			return t
		}
	}
	return nil
}

type Sproxy struct {
}

func (h *Sproxy) OnNetMade(t *lbbnet.Transport) {
	fmt.Println("s made net")
	SM.AddServer(t)
}

func (h *Sproxy) OnNetLost(t *lbbnet.Transport) {
	fmt.Println("s lost net")
	SM.RemoveServer(t)
}

func (h *Sproxy) OnNetData(data *lbbnet.NetPacket) {
	id := SM.GetService(data.Rw)
	data.ServerId = uint32(id)

	client := CM.GetClient(data.UserId)
	if client == nil {
		fmt.Println("get client emtpy")
		return
	}
	client.WriteData(data.Serialize())
}

type CManager struct {
	clients []*lbbnet.Transport
	cs      map[string]*lbbnet.TClient
	sync.Mutex
}

func newCManager() *CManager {
	return &CManager{cs: make(map[string]*lbbnet.TClient)}
}

func (c *CManager) GetClient(sharding uint64) *lbbnet.Transport {
	c.Lock()
	defer c.Unlock()
	if len(c.clients) == 0 {
		return nil
	}
	fmt.Println("len client", len(c.clients))
	index := sharding % uint64(len(c.clients))
	return c.clients[index]
}

func (c *CManager) AddClient(t *lbbnet.Transport) {
	c.Lock()
	defer c.Unlock()
	c.clients = append(c.clients, t)
}

func (c *CManager) AddTClient(addr string, t *lbbnet.TClient) {
	c.Lock()
	defer c.Unlock()
	c.cs[addr] = t
}

func (c *CManager) RemClient(t *lbbnet.Transport) {
	fmt.Println("---------lbbnet------remove client")
	c.Lock()
	defer c.Unlock()
	for index := range c.clients {
		if c.clients[index] == t {
			c.clients = append(c.clients[0:index], c.clients[index+1:]...)
		}
	}
	addr := t.RemoteAddr()
	tclient := c.cs[addr]
	if tclient == nil {
		fmt.Println("---------lbbnet------remove client empty", addr)
		return
	}
	tclient.Close()
	delete(c.cs, addr)
}

func (c *CManager) RemoveServerByAddr(addr string) {
	fmt.Println("---------consul------remove client", addr)
	c.Lock()
	defer c.Unlock()
	t := c.cs[addr]
	if t == nil {
		fmt.Println("-------consul--------remove client empty")
		return
	}
	delete(c.cs, addr)
	for index := range c.clients {
		if c.clients[index].RemoteAddr() == addr {
			c.clients = append(c.clients[0:index], c.clients[index+1:]...)
		}
	}
	t.Close()
}

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *lbbnet.Transport) {
	fmt.Println("c made net")
	CM.AddClient(t)
}

func (h *Cproxy) OnNetLost(t *lbbnet.Transport) {
	fmt.Println("c lost net")
	CM.RemClient(t)
}

func (h *Cproxy) OnNetData(data *lbbnet.NetPacket) {
	s := SM.GetServiceById(data.ServerId)
	if s == nil {
		return
	}
	s.WriteData(data.Serialize())
}
