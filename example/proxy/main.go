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

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 9999, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "client_id_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "client_id", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:8888", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "serverNode_1", "found server name")
	flag.Parse()
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		fmt.Println("open return", err)
		return
	}

	hello := &Hello{}
	hello.init()
	s, err := lbbnet.NewTServer(":9099", hello, 2*time.Second)
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
			for k, v := range services {
				if _, ok := oldSer[k]; !ok {
					go func(s *lbbconsul.ServiceInfo) {
						sendData(s)
					}(v)
				}
			}
			oldSer = services
		}
	}()

	<-exist
	lbbconsul.GConsulClient.Close()
}

func compareDiff(old, new map[string]*lbbconsul.ServiceInfo) (rem, update, add map[string]**lbbconsul.ServiceInfo) {
	var rem1, update1, add1 []**lbbconsul.ServiceInfo
	for k, v := range old {
		if v2, ok := new[k]; ok {
			if v2.Ip == v.Ip && v2.Port == v.Port {
			} else {
				update1 = append(update1, v2)
			}
		} else {
			rem1 = append(rem1, v)
		}
	}

	for k, v := range new {
		if _, ok := old[k]; !ok {
			add1 = append(add1, v)
		}
	}
	if len(rem1) > 0 {
		rem = make(map[string]**lbbconsul.ServiceInfo)
		for index := range rem1 {
			rem[rem1[index].ServiceID] = rem1[index]
		}
	}
	if len(update1) > 0 {
		update = make(map[string]**lbbconsul.ServiceInfo)
		for index := range update1 {
			rem[update1[index].ServiceID] = update1[index]
		}
	}
	if len(add1) > 0 {
		add = make(map[string]**lbbconsul.ServiceInfo)
		for index := range add1 {
			rem[add1[index].ServiceID] = add1[index]
		}
	}
	return
}

var SM *ServerManager

type ServerManager struct {
	seq     int
	clients map[int]*lbbnet.Transport
	sync.Mutex
}

func newServerManager() {
	return &ServerManager{0, make(map[int]*lbbnet.Transport)}
}
func (s *ServerManager) RemoveByTransport(t *lbbnet.Transport) {
	s.Lock()
	defer s.Unlock()
	for id, tmp := range s.clients {
		if tmp == t {
			delete(s.clients, id)
			break
		}
	}
}

func (s *ServerManager) Remove(id int) {
	s.Lock()
	defer s.Unlock()
	delete(s.clients, id)
}

func (s *ServerManager) SetServiceId(id int, conn *lbbnet.Transport) {
	s.Lock()
	defer s.Unlock()
	s.clients[id] = conn
}

func (s *ServerManager) GetServiceId() int {
	s.Lock()
	defer s.Unlock()
	s.seq++
	return s.seq
}

func (s *ServerManager) GetService(sid int) **lbbnet.Transport {
	s.Lock()
	defer s.Unlock()
	return s.clients[sid]
}

type Sproxy struct {
	seqid int
}

func (h *Sproxy) OnNetMade(t *lbbnet.Transport) {
	id := SM.GetServiceId()
	h.seqid = id
	SM.SetServiceId(id, t)
}

func (h *Sproxy) OnNetLost(t *lbbnet.Transport) {
	SM.RemoveByTransport(t)
}

func (h *Sproxy) OnNetData(data *lbbnet.NetPacket) {
	data.ServerId = uint32(h.seqid)

	// client[data.UserId%10].Send(data)
}

type CM struct {
	num     int
	clients []*lbbnet.Transport
}

func (c *CM) AddClient(t *lbbNet.Transport) {
	c.clients = append(c.clients, t)
}

func (c *CM) RemClient(t *lbbNet.Transport) {
	for index := range c.clients {
		if c.clients[index] == t {
			c.clients = append(c.clients[0:index], c.clients[index+1:]...)
		}
	}
}

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *lbbnet.Transport) {

}

func (h *Cproxy) OnNetLost(t *lbbnet.Transport) {
}

func (h *Cproxy) OnNetData(data *lbbnet.NetPacket) {
}

func sendData(service *lbbconsul.ServiceInfo) (lbbnet.TClient, error) {
	t, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", service.Ip, service.Port), pdf, 3*time.Second)
}

func closeClient(sid string) {
}

type Hello struct {
	close bool
}

func (h *Hello) OnNetMade(t *lbbnet.Transport) {
	fmt.Println("connect mad")
	go func() {
		for i := 0; i < 1000; i++ {
			time.Sleep(100 * time.Millisecond)
			p := &lbbnet.NetPacket{UserId: 123, ServerId: 1, SessionId: 2, PacketType: uint32(1 + i%2)}
			data := p.Encoder([]byte(fmt.Sprintf("client%d", i)))
			t.WriteData(data)
		}
	}()
}

func (h *Hello) OnNetData(data *lbbnet.NetPacket) {
	SM.GetService(data.ServerId)
}
func (h *Hello) OnNetLost(t *lbbnet.Transport) {
	fmt.Println("connect lost")
}

func (h *Hello) Close() {
	h.close = true
	fmt.Println("hello close")
}
