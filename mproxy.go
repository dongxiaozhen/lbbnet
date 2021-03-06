package lbbnet

import (
	"encoding/json"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

type MSproxy struct {
	cm         *ClientManager
	psm        *PServerManager
	ServerInfo []byte
	ServerId   string
}

func NewMSproxy(cm *ClientManager, psm *PServerManager, serverId string, serverInfo []byte) *MSproxy {
	s := &MSproxy{cm: cm, psm: psm, ServerId: serverId, ServerInfo: serverInfo}
	return s
}

func (h *MSproxy) reverseRegisterService() {
	log.Warn("MSP--------->reverseRegister")
	for _, t := range h.cm.GetClients() {
		p := &NetPacket{PacketType: PTypeSysNotifyServicesChange, ReqType: MTypeOneWay}
		err := t.WriteData(p)
		if err != nil {
			log.Error("reverse register server err: s=", t.RemoteAddr(), err)
		} else {
			log.Warn("reverse register server s=", t.RemoteAddr())
		}
	}
}

func (h *MSproxy) registerService(t *Transport) error {
	p1 := &NetPacket{PacketType: PTypeSysNotifyServerId, ReqType: MTypeOneWay, Data: []byte(h.ServerId)}
	p := &NetPacket{PacketType: PTypeSysObtainServices, ReqType: MTypeCall}
	t.WriteData(p1)
	return t.WriteData(p)
}
func (h *MSproxy) OnNetMade(t *Transport) {
	log.Warn("MSP--------->made", t.RemoteAddr())
	err := h.registerService(t)
	if err != nil {
		log.Error("send server register", t.RemoteAddr())
	}
}

func (h *MSproxy) OnNetLost(t *Transport) {
	log.Warn("MSP--------->lost", t.RemoteAddr())
	h.psm.RemServer(t)
	h.reverseRegisterService()
}

func (h *MSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == PTypeSysObtainServices && data.ReqType == MTypeReply {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("MSP server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("MSP server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		h.psm.AddServer(data.Rw, ids)
		// 反向注册
		h.reverseRegisterService()
		return
	} else if data.ReqType == MTypeRoute {
		buf := make([]byte, 0, len(h.ServerInfo)+len(data.Data))
		buf = append(buf, h.ServerInfo...)
		buf = append(buf, data.Data...)
		data.Data = buf
	}

	s := h.cm.GetClientById(data.From2)
	if s == nil {
		log.Warn("MSproxy client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}

type MCproxy struct {
	cm  *ClientManager
	psm *PServerManager
}

func NewMCproxy(cm *ClientManager, psm *PServerManager) *MCproxy {
	c := &MCproxy{cm, psm}
	return c
}

func (h *MCproxy) OnNetMade(t *Transport) {
	log.Debug("MCP-------------s made net")
}

func (h *MCproxy) OnNetLost(t *Transport) {
	log.Debug("MCP-------------s lost net")
	h.cm.RemoveClient(t)
}

func (h *MCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref(lbbconsul.GetConsulServerId()).Deref()

	if data.PacketType == PTypeSysNotifyServerId && data.ReqType == MTypeOneWay {
		data.Rw.SetRemoteId(string(data.Data))
		h.cm.AddClient(data.Rw)
		log.Warn("add client notifyserver", string(data.Data))
		return
	}
	if data.PacketType == PTypeSysObtainServices && data.ReqType == MTypeCall {
		log.Warn("MCP get register packet", data.Rw.RemoteAddr())
		h.psm.GetServerIds(data)
		return
	}
	id := h.cm.GetClient(data.Rw)
	data.From2 = uint32(id)

	client := h.psm.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("MCP get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}
