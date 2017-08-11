package lbbnet

import (
	"fmt"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *Transport) {
	log.Debug("CP made net ", t.RemoteAddr())
	CM.AddClient(t)
}

func (h *Cproxy) OnNetLost(t *Transport) {
	log.Debug("CP lost net", t.RemoteAddr())
	CM.RemoveClient(t)
}

func (h *Cproxy) OnNetData(data *NetPacket) {
	id := CM.GetClient(data.Rw)
	data.From2 = uint32(id)

	defer goref.Ref("proxy").Deref()

	client := SM.GetServer(data.UserId)
	if client == nil {
		log.Warn("Cp get client emtpy")
		return
	}
	client.WriteData(data)
}

type Sproxy struct {
}

func (h *Sproxy) OnNetMade(t *Transport) {
	log.Warn("SP made---------> ", t.RemoteAddr())
	SM.AddServer(t)
}

func (h *Sproxy) OnNetLost(t *Transport) {
	log.Warn("SP lost---------> ", t.RemoteAddr())
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
				log.Warn(" CompareDiff remvoe server------> ", *v)
				pp.RemoveServerByAddr(addr)
				t, err := NewTClient(addr, pf, 0)
				if err != nil {
					log.Warn("CompareDiff proxy server err ", addr, err)
				} else {
					pp.AddTServer(addr, t)
				}
			}
		} else {
			log.Debug("CompareDiff remove server ---> ", *v)
			pp.TmpRemoveServerByAddr(addr)
		}
	}

	for k, v := range new {
		addr := fmt.Sprintf("%s:%d", v.IP, v.Port)
		if _, ok := old[k]; !ok {
			log.Debug("CompareDiff add server ", *v)
			t, err := NewTClient(addr, pf, 0)
			if err != nil {
				log.Warn("CompareDiff proxy client err", addr, err)
			} else {
				pp.AddTServer(addr, t)
			}
		}
	}
}
