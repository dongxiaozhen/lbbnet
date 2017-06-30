package lbbnet

import (
	"fmt"
	"sync"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
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
	log.Debug("len client", len(c.clients))
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
	c.Lock()
	defer c.Unlock()
	index := -1
	for i := range c.clients {
		if c.clients[i] == t {
			index = i
			break
		}
	}
	if index != -1 {
		c.clients = append(c.clients[:index], c.clients[index+1:]...)
	}

	addr := t.RemoteAddr()
	tclient := c.cs[addr]
	if tclient == nil {
		log.Warn("---------lbbnet------remove client empty", addr)
		return
	}
	tclient.Close()
	delete(c.cs, addr)
}

func (c *CManager) RemoveServerByAddr(addr string) {
	log.Debug("---------consul------remove client", addr)
	c.Lock()
	defer c.Unlock()
	t := c.cs[addr]
	if t == nil {
		return
	}
	delete(c.cs, addr)
	i := -1
	for index := range c.clients {
		if c.clients[index].RemoteAddr() == addr {
			i = index
			break
		}
	}
	if i != -1 {
		c.clients = append(c.clients[:i], c.clients[i+1:]...)
	}
	t.Close()
}

type Sproxy struct {
}

func (h *Sproxy) OnNetMade(t *Transport) {
	log.Debug("-------------s made net")
	SM.AddServer(t)
}

func (h *Sproxy) OnNetLost(t *Transport) {
	log.Debug("-------------s lost net")
	SM.RemoveServer(t)
}

func (h *Sproxy) OnNetData(data *NetPacket) {
	id := SM.GetService(data.Rw)
	data.ServerId = uint32(id)

	defer goref.Ref("proxy").Deref()

	client := CM.GetClient(data.UserId)
	if client == nil {
		log.Debug("get client emtpy")
		return
	}
	client.WriteData(data)
}

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *Transport) {
	log.Debug("c made net")
	CM.AddClient(t)
}

func (h *Cproxy) OnNetLost(t *Transport) {
	log.Debug("c lost net")
	CM.RemClient(t)
}

func (h *Cproxy) OnNetData(data *NetPacket) {
	s := SM.GetServiceById(data.ServerId)
	if s == nil {
		return
	}
	s.WriteData(data)
}

func CompareDiff(old, new map[string]*lbbconsul.ServiceInfo, pf Protocol) {
	for k, v := range old {
		if v2, ok := new[k]; ok {
			if v2.IP == v.IP && v2.Port == v.Port {
			} else {
				log.Debug("-------------------remvoe server", *v)
				CM.RemoveServerByAddr(fmt.Sprintf("%s:%d", v.IP, v.Port))
				t, err := NewTClient(fmt.Sprintf("%s:%d", v2.IP, v2.Port), pf, 60*time.Second)
				if err != nil {
					log.Warn("proxy client err", err)
				} else {
					CM.AddTClient(fmt.Sprintf("%s:%d", v2.IP, v2.Port), t)
				}
			}
		} else {
			log.Debug("-------------------remove server", *v)
			CM.RemoveServerByAddr(fmt.Sprintf("%s:%d", v.IP, v.Port))
		}
	}

	for k, v := range new {
		if _, ok := old[k]; !ok {
			log.Debug("-------------------add server", *v)
			t, err := NewTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), pf, 60*time.Second)
			if err != nil {
				log.Warn("proxy client err", err)
			} else {
				CM.AddTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), t)
			}
		}
	}
}
