package lbbnet

import (
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

type Cproxy struct {
}

func (h *Cproxy) OnNetMade(t *Transport) {
	log.Debug("CP made net ", t.RemoteAddr())
}

func (h *Cproxy) OnNetLost(t *Transport) {
	log.Debug("CP lost net", t.RemoteAddr())
	CM.RemoveClient(t)
}

func (h *Cproxy) OnNetData(data *NetPacket) {
	defer goref.Ref("proxy").Deref()

	if data.PacketType == PTypeNotifyServer && data.ReqType == MTypeOneWay {
		data.Rw.SetRemoteId(string(data.Data))
		CM.AddClient(data.Rw)
		return
	}

	id := CM.GetClient(data.Rw)
	data.From2 = uint32(id)

	client := SM.GetServer(data.UserId)
	if client == nil {
		log.Warn("Cp get client emtpy")
		return
	}
	client.WriteData(data)
}

type Sproxy struct {
	ServerInfo []byte
	ServerId   string
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

func (h *Sproxy) registerServiceId(t *Transport) error {
	p1 := &NetPacket{PacketType: PTypeNotifyServer, ReqType: MTypeOneWay, Data: []byte(h.ServerId)}
	return t.WriteData(p1)
}

func (h *Sproxy) OnNetMade(t *Transport) {
	log.Warn("SP made---------> ", t.RemoteAddr())
	SM.AddServer(t)
	err := h.registerServiceId(t)
	if err != nil {
		log.Error("SP send server register", t.RemoteAddr())
	}
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
		log.Debug("SP discard -->", data.PacketType, data.SeqId, string(data.Data))
		return
	}
	s.WriteData(data)
}
