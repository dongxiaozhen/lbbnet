package lbbnet

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

var PSM = NewPServerManager()

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
	for id := range c.clients {
		s = append(s, id)
	}

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
	log.Warn("PSM---------Add", t.RemoteAddr())
	c.Lock()
	defer c.Unlock()
	c.ids[t.RemoteAddr()] = ids
	for _, id := range ids {
		c.clients[id] = append(c.clients[id], t)
	}
}

func (c *PServerManager) AddTServer(addr string, t *TClient) {
	log.Warn("PSM---------AddT", addr)
	c.Lock()
	defer c.Unlock()
	c.cs[addr] = t
}

// 被动关闭一个后端服务
func (c *PServerManager) RemServer(t *Transport) {
	log.Warn("PSM---------服务断开连接", t.RemoteAddr())
	c.Lock()
	defer c.Unlock()
	addr := t.RemoteAddr()
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
			}
		}
	}

	tclient := c.cs[addr]
	if tclient == nil {
		log.Warn("---------lbbnet------remove client empty", addr)
		return
	}
	tclient.Close()
	delete(c.cs, addr)
	delete(c.ids, addr)
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
	log.Warn("PSM---------代理主动关闭连接", addr)
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

type PSproxy struct {
}

func (h *PSproxy) registerService(t *Transport) error {
	p := &NetPacket{PacketType: 0}
	return t.WriteData(p)
}
func (h *PSproxy) OnNetMade(t *Transport) {
	log.Warn("PSP--------->made", t.RemoteAddr())
	err := h.registerService(t)
	if err != nil {
		log.Error("send server register", t.RemoteAddr())
	}
}

func (h *PSproxy) OnNetLost(t *Transport) {
	log.Warn("PSP--------->lost", t.RemoteAddr())
	PSM.RemServer(t)
}

func (h *PSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == 0 {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		return
	}
	s := CM.GetClientById(data.From2)
	if s == nil {
		log.Warn("PSproxy client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}

type PCproxy struct {
}

func (h *PCproxy) OnNetMade(t *Transport) {
	log.Debug("PCP-------------s made net")
	CM.AddClient(t)
}

func (h *PCproxy) OnNetLost(t *Transport) {
	log.Debug("PCP-------------s lost net")
	CM.RemoveClient(t)
}

func (h *PCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref("Pproxy").Deref()

	if data.PacketType == 0 {
		PSM.GetServerIds(data)
		return
	}
	id := CM.GetClient(data.Rw)
	data.From2 = uint32(id)

	client := PSM.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("pcp get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}

type PpCproxy struct {
}

func (h *PpCproxy) OnNetMade(t *Transport) {
	log.Debug("PCP-------------s made net")
	CM.AddClient(t)
}

func (h *PpCproxy) OnNetLost(t *Transport) {
	log.Debug("PCP-------------s lost net")
	CM.RemoveClient(t)
}

func (h *PpCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref("Pproxy").Deref()

	if data.PacketType == 0 {
		PSM.GetServerIds(data)
		return
	}
	id := CM.GetClient(data.Rw)
	data.From1 = uint32(id)

	client := PSM.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("pcp get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}

type PpSproxy struct {
}

func (h *PpSproxy) registerService(t *Transport) error {
	p := &NetPacket{PacketType: 0}
	return t.WriteData(p)
}
func (h *PpSproxy) OnNetMade(t *Transport) {
	log.Warn("PpSP--------->made", t.RemoteAddr())
	err := h.registerService(t)
	if err != nil {
		log.Error("send server register", t.RemoteAddr())
	}
}

func (h *PpSproxy) OnNetLost(t *Transport) {
	log.Warn("PpSP--------->lost", t.RemoteAddr())
	PSM.RemServer(t)
}

func (h *PpSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == 0 {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		return
	}
	s := CM.GetClientById(data.From1)
	if s == nil {
		log.Warn("PpSproxy client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}
