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
				log.Warn("PSM add type-->", id, addr)
				c.ids[addr] = append(c.ids[addr], id)
			}
		}

	} else {
		for _, id := range ids {
			c.clients[id] = append(c.clients[id], t)
			log.Warn("PSM add type-->", id, addr)
		}
		c.ids[addr] = ids
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

func (h *PSproxy) reverseRegisterService() {
	log.Warn("PSP--------->reverseRegister")
	for _, t := range CM.GetClients() {
		p := &NetPacket{PacketType: PTypeReverseRegistServer}
		err := t.WriteData(p)
		if err != nil {
			log.Error("reverse register server err: s=", t.RemoteAddr(), err)
		} else {
			log.Warn("reverse register server s=", t.RemoteAddr())
		}
	}
}

func (h *PSproxy) registerService(t *Transport) error {
	p := &NetPacket{PacketType: PTypeRegistServer}
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
	h.reverseRegisterService()
}

func (h *PSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == PTypeRegistServer {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		// 反向注册
		h.reverseRegisterService()
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

	if data.PacketType == PTypeRegistServer {
		log.Warn("get register packet", data.Rw.RemoteAddr())
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

	if data.PacketType == PTypeRegistServer {
		log.Warn("PpCp remote get Regist-->", data.Rw.RemoteAddr())
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
	p := &NetPacket{PacketType: PTypeRegistServer}
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
	if data.PacketType == PTypeRegistServer {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		return
	} else if data.PacketType == PTypeReverseRegistServer {
		err := h.registerService(data.Rw)
		if err != nil {
			log.Error("reverse server regist response  register err", data.Rw.RemoteAddr(), err)
		}
		return
	}
	s := CM.GetClientById(data.From1)
	if s == nil {
		log.Warn("PpSproxy client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}
