package lbbnet

import (
	"fmt"
	"sync"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

var CM = NewClientManager()
var SM = NewServerManager()

type ClientManager struct {
	seq     uint32
	clients map[*Transport]uint32
	sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		seq:     0,
		clients: make(map[*Transport]uint32),
	}
}

func (s *ClientManager) RemoveClient(t *Transport) {
	s.Lock()
	defer s.Unlock()
	t.Close()
	delete(s.clients, t)
}

func (s *ClientManager) AddClient(conn *Transport) {
	s.Lock()
	defer s.Unlock()
	s.seq++
	s.clients[conn] = s.seq
}

func (s *ClientManager) GetClient(t *Transport) uint32 {
	s.RLock()
	defer s.RUnlock()
	return s.clients[t]
}

func (s *ClientManager) GetClientById(sid uint32) *Transport {
	s.RLock()
	defer s.RUnlock()
	for t, id := range s.clients {
		if id == sid {
			return t
		}
	}
	return nil
}

func (s *ClientManager) GetClients() []*Transport {
	s.RLock()
	defer s.RUnlock()
	ret := make([]*Transport, 0, len(s.clients))
	for t := range s.clients {
		ret = append(ret, t)
	}
	return ret
}

type ServerManager struct {
	clients []*Transport
	cs      map[string]*TClient
	sync.RWMutex
}

func NewServerManager() *ServerManager {
	return &ServerManager{cs: make(map[string]*TClient)}
}

func (c *ServerManager) HasServer(addr string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.cs[addr]
	return ok
}

func (c *ServerManager) GetServer(sharding uint64) *Transport {
	c.RLock()
	defer c.RUnlock()
	if len(c.clients) == 0 {
		return nil
	}
	index := sharding % uint64(len(c.clients))
	log.Warn("user= ", sharding, "proxy-> ", c.clients[index].RemoteAddr(), "now client_len=", len(c.clients))
	return c.clients[index]
}

func (c *ServerManager) AddServer(t *Transport) {
	log.Warn("SM---------Add", t.RemoteAddr())
	c.Lock()
	defer c.Unlock()
	c.clients = append(c.clients, t)
}

func (c *ServerManager) AddTServer(addr string, t *TClient) {
	log.Warn("SM---------AddT", addr)
	c.Lock()
	defer c.Unlock()
	c.cs[addr] = t
}

// 被动关闭一个后端服务
func (c *ServerManager) RemServer(t *Transport) {
	log.Warn("SM---------服务断开连接", t.RemoteAddr())
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

//主动关闭一个后端服务--暂时
func (c *ServerManager) TmpRemoveServerByAddr(addr string) {
	log.Warn("SM---------暂停服务", addr)
	c.Lock()
	defer c.Unlock()
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
}

// 主动关闭一个后端服务
func (c *ServerManager) RemoveServerByAddr(addr string) {
	log.Warn("SM---------代理主动关闭连接", addr)
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

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *Transport) {
	log.Debug("CP-------------s made net")
	CM.AddClient(t)
}

func (h *Cproxy) OnNetLost(t *Transport) {
	log.Debug("CP-------------s lost net")
	CM.RemoveClient(t)
}

func (h *Cproxy) OnNetData(data *NetPacket) {
	id := CM.GetClient(data.Rw)
	data.From2 = uint32(id)

	defer goref.Ref("proxy").Deref()

	client := SM.GetServer(data.UserId)
	if client == nil {
		log.Warn("get client emtpy")
		return
	}
	client.WriteData(data)
}

type Sproxy struct {
}

func (h *Sproxy) OnNetMade(t *Transport) {
	log.Warn("SP--------->made", t.RemoteAddr())
	SM.AddServer(t)
}

func (h *Sproxy) OnNetLost(t *Transport) {
	log.Warn("SP--------->lost", t.RemoteAddr())
	SM.RemServer(t)
}

func (h *Sproxy) OnNetData(data *NetPacket) {
	s := CM.GetClientById(data.From2)
	if s == nil {
		return
	}
	s.WriteData(data)
}

func CompareDiff(old, new map[string]*lbbconsul.ServiceInfo, pf Protocol, pp PProtocol) {
	for k, v := range old {
		addr := fmt.Sprintf("%s:%d", v.IP, v.Port)
		if v2, ok := new[k]; ok {
			if v2.IP != v.IP || v2.Port != v.Port || (v2.IP == v.IP && v2.Port == v.Port && !pp.HasServer(addr)) {
				log.Debug("-------------------remvoe server", *v)
				pp.RemoveServerByAddr(addr)
				t, err := NewTClient(addr, pf, 0)
				if err != nil {
					log.Warn("proxy servererr", addr, err)
				} else {
					pp.AddTServer(addr, t)
				}
			}
		} else {
			log.Debug("-------------------remove server", *v)
			pp.TmpRemoveServerByAddr(addr)
		}
	}

	for k, v := range new {
		addr := fmt.Sprintf("%s:%d", v.IP, v.Port)
		if _, ok := old[k]; !ok {
			log.Debug("-------------------add server", *v)
			t, err := NewTClient(addr, pf, 60*time.Second)
			if err != nil {
				log.Warn("proxy client err", addr, err)
			} else {
				pp.AddTServer(addr, t)
			}
		}
	}
}
