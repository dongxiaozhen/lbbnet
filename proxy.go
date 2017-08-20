package lbbnet

import (
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
	ServerInfo []byte
}

func (h *Sproxy) reverseRegisterService() {
	log.Warn("SP--------->reverseRegister")
	for _, t := range CM.GetClients() {
		p := &NetPacket{PacketType: PTypeReverseRegistServer, ReqType: MTypeOneWay}
		err := t.WriteData(p)
		if err != nil {
			log.Error("SP reverse register server err: s=", t.RemoteAddr(), err)
		} else {
			log.Warn("SP reverse register server s=", t.RemoteAddr())
		}
	}
}

func (h *Sproxy) OnNetMade(t *Transport) {
	log.Warn("SP made---------> ", t.RemoteAddr())
	SM.AddServer(t)
	h.reverseRegisterService()
}

func (h *Sproxy) OnNetLost(t *Transport) {
	log.Warn("SP lost---------> ", t.RemoteAddr())
	SM.RemServer(t)
	h.reverseRegisterService()
}

func (h *Sproxy) OnNetData(data *NetPacket) {
	if data.ReqType == MTypeRoute {
		buf := make([]byte, 0, len(h.ServerInfo)+len(data.Data))
		buf = append(buf, h.ServerInfo...)
		buf = append(buf, data.Data...)
		data.Data = buf
	}
	s := CM.GetClientById(data.From2)
	if s == nil {
		return
	}
	s.WriteData(data)
}
