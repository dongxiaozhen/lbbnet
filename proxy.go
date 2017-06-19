package lbbnet

import (
	"fmt"
	"sync"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
)

var SM = NewServerManager()
var CM = NewCManager()

type ServerManager struct {
	seq     uint32
	clients map[*Transport]uint32
	sync.Mutex
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		seq:     0,
		clients: make(map[*Transport]uint32),
	}
}

func (s *ServerManager) RemoveServer(t *Transport) {
	s.Lock()
	defer s.Unlock()
	t.Close()
	delete(s.clients, t)
}

func (s *ServerManager) AddServer(conn *Transport) {
	s.Lock()
	defer s.Unlock()
	s.seq++
	s.clients[conn] = s.seq
}

func (s *ServerManager) GetService(t *Transport) uint32 {
	s.Lock()
	defer s.Unlock()
	return s.clients[t]
}

func (s *ServerManager) GetServiceById(sid uint32) *Transport {
	s.Lock()
	defer s.Unlock()
	for t, id := range s.clients {
		if id == sid {
			return t
		}
	}
	return nil
}

type CManager struct {
	clients []*Transport
	cs      map[string]*TClient
	sync.Mutex
}

func NewCManager() *CManager {
	return &CManager{cs: make(map[string]*TClient)}
}

func (c *CManager) GetClient(sharding uint64) *Transport {
	c.Lock()
	defer c.Unlock()
	if len(c.clients) == 0 {
		return nil
	}
	fmt.Println("len client", len(c.clients))
	index := sharding % uint64(len(c.clients))
	return c.clients[index]
}

func (c *CManager) AddClient(t *Transport) {
	c.Lock()
	defer c.Unlock()
	c.clients = append(c.clients, t)
}

func (c *CManager) AddTClient(addr string, t *TClient) {
	c.Lock()
	defer c.Unlock()
	c.cs[addr] = t
}

func (c *CManager) RemClient(t *Transport) {
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

type Sproxy struct {
}

func (h *Sproxy) OnNetMade(t *Transport) {
	fmt.Println("s made net")
	SM.AddServer(t)
}

func (h *Sproxy) OnNetLost(t *Transport) {
	fmt.Println("s lost net")
	SM.RemoveServer(t)
}

func (h *Sproxy) OnNetData(data *NetPacket) {
	id := SM.GetService(data.Rw)
	data.ServerId = uint32(id)

	client := CM.GetClient(data.UserId)
	if client == nil {
		fmt.Println("get client emtpy")
		return
	}
	client.WriteData(data.Serialize())
}

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *Transport) {
	fmt.Println("c made net")
	CM.AddClient(t)
}

func (h *Cproxy) OnNetLost(t *Transport) {
	fmt.Println("c lost net")
	CM.RemClient(t)
}

func (h *Cproxy) OnNetData(data *NetPacket) {
	s := SM.GetServiceById(data.ServerId)
	if s == nil {
		return
	}
	s.WriteData(data.Serialize())
}

func CompareDiff(old, new map[string]*lbbconsul.ServiceInfo, pf Protocol) {
	for k, v := range old {
		if v2, ok := new[k]; ok {
			if v2.IP == v.IP && v2.Port == v.Port {
			} else {
				fmt.Println("-------------------remvoe server", *v)
				CM.RemoveServerByAddr(fmt.Sprintf("%s:%d", v.IP, v.Port))
				t, err := NewTClient(fmt.Sprintf("%s:%d", v2.IP, v2.Port), pf, 60*time.Second)
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
			t, err := NewTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), pf, 60*time.Second)
			if err != nil {
				fmt.Println("proxy client err", err)
			} else {
				CM.AddTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), t)
			}
		}
	}
}
