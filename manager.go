package lbbnet

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dongxiaozhen/lbbref/goref"
	"github.com/dongxiaozhen/lbbsort"
	log "github.com/donnie4w/go-logger/logger"
)

var PSM = NewPServerManager()
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

type PServerManager struct {
	clients map[uint32][]*Transport
	cs      map[string]*TClient
	ids     map[string][]uint32
	sync.RWMutex
}

func NewPServerManager() *PServerManager {
	return &PServerManager{cs: make(map[string]*TClient), clients: make(map[uint32][]*Transport), ids: make(map[string][]uint32)}
}

func (c *PServerManager) GetServerIds(data *NetPacket) {
	c.RLock()
	defer c.RUnlock()
	s := make([]uint32, 0, len(c.clients))
	for id, value := range c.clients {
		if len(value) > 0 {
			s = append(s, id)
		}
	}
	lbbsort.Uint32s(s)

	var err error
	data.Data, err = json.Marshal(s)
	if err != nil {
		log.Error("marshal server id err", err)
		data.Rw.WriteData(data)
		return
	}
	log.Warn("regist server id", data.Rw.RemoteAddr(), string(data.Data))
	data.Rw.WriteData(data)
}

func (c *PServerManager) HasServer(addr string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.cs[addr]
	return ok
}

func (c *PServerManager) GetServer(sharding uint64, packId uint32) *Transport {
	c.RLock()
	defer c.RUnlock()
	slen := len(c.clients[packId])
	if slen == 0 {
		return nil
	}
	goref.Ref(fmt.Sprintf("packId_%d", packId)).Deref()
	index := sharding % uint64(slen)
	log.Debug("user= ", sharding, "proxy-> ", c.clients[packId][index].RemoteAddr(), "now client_len=", slen)
	return c.clients[packId][index]
}

func (c *PServerManager) AddServer(t *Transport, ids []uint32) {
	log.Warn("PSM---------Add", t.RemoteAddr(), ids)
	c.Lock()
	defer c.Unlock()
	addr := t.RemoteAddr()
	if tids, ok := c.ids[addr]; ok {

		// 删除
		log.Warn("delete from PSM")
		lbbsort.Uint32s(tids)
		lenIds := len(ids)
		for _, id := range tids {
			sids := lbbsort.Uint32Slice(ids)
			index := sids.Search(id)
			if index >= lenIds || ids[index] != id {
				for tmp, trp := range c.clients[id] {
					if trp == t {
						c.clients[id] = append(c.clients[id][:tmp], c.clients[id][tmp+1:]...)
						log.Warn("PSM remove type-->", id, addr)
					}
				}
				for tmp, id2 := range c.ids[addr] {
					if id2 == id {
						c.ids[addr] = append(c.ids[addr][:tmp], c.ids[addr][tmp+1:]...)
					}
				}
			}
		}

		log.Warn("Add to PSM")
		// 添加
		tids = c.ids[addr]
		lbbsort.Uint32s(tids)
		lenTids := len(tids)
		for _, id := range ids {
			sids := lbbsort.Uint32Slice(tids)
			index := sids.Search(id)
			if index >= lenTids || tids[index] != id {
				c.clients[id] = append(c.clients[id], t)
				log.Warn("PSM add type--> ", id, addr)
				c.ids[addr] = append(c.ids[addr], id)
			}
		}

	} else {
		for _, id := range ids {
			c.clients[id] = append(c.clients[id], t)
			log.Warn("PSM add type--> ", id, addr)
		}
		c.ids[addr] = ids
	}
}

func (c *PServerManager) AddTServer(addr string, t *TClient) {
	log.Warn("PSM---------AddT ", addr)
	c.Lock()
	defer c.Unlock()
	c.cs[addr] = t
}

// 被动关闭一个后端服务
func (c *PServerManager) RemServer(t *Transport) {
	c.Lock()
	defer c.Unlock()
	addr := t.RemoteAddr()
	log.Warn("PSM---------服务断开连接", addr)
	index := -1
	if s, ok := c.ids[addr]; ok {
		for _, id := range s {
			index = -1
			for i := range c.clients[id] {
				if c.clients[id][i] == t {
					index = i
					break
				}
			}
			if index != -1 {
				c.clients[id] = append(c.clients[id][:index], c.clients[id][index+1:]...)
				log.Warn("PSM  remove type-->", id, addr)
			} else {
				log.Warn("PSM  remove type not find-->", id, addr)
			}
		}
	} else {
		log.Warn("PSM  RemServer empty ids", addr)
	}

	delete(c.ids, addr)

	tclient := c.cs[addr]
	if tclient == nil {
		log.Warn("---------lbbnet------remove client empty", addr)
		return
	}
	delete(c.cs, addr)
	tclient.Close()
}

//主动关闭一个后端服务--暂时
func (c *PServerManager) TmpRemoveServerByAddr(addr string) {
	log.Warn("PSM---------暂停服务", addr)
	c.Lock()
	defer c.Unlock()
	var (
		s     []uint32
		ok    bool
		index = -1
	)

	if s, ok = c.ids[addr]; !ok {
		return
	}
	for _, id := range s {
		index = -1
		for i := range c.clients[id] {
			if c.clients[id][i].RemoteAddr() == addr {
				index = i
				break
			}
		}
		if index != -1 {
			c.clients[id] = append(c.clients[id][:index], c.clients[id][index+1:]...)
		}
	}
}

// 主动关闭一个后端服务
func (c *PServerManager) RemoveServerByAddr(addr string) {
	log.Warn("PSM---------代理主动关闭连接 ", addr)
	c.Lock()
	defer c.Unlock()
	t := c.cs[addr]
	if t == nil {
		return
	}
	delete(c.cs, addr)

	index := -1
	if s, ok := c.ids[addr]; ok {
		for _, id := range s {
			index = -1
			for i := range c.clients[id] {
				if c.clients[id][i].RemoteAddr() == addr {
					index = i
					break
				}
			}
			if index != -1 {
				c.clients[id] = append(c.clients[id][:index], c.clients[id][index+1:]...)
			}
		}
	}
	delete(c.ids, addr)

	t.Close()
}
