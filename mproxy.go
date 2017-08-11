package lbbnet

import (
	"encoding/json"

	log "github.com/donnie4w/go-logger/logger"
	"github.com/mreithub/goref"
)

type MSproxy struct {
}

func (h *MSproxy) reverseRegisterService() {
	log.Warn("MSP--------->reverseRegister")
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

func (h *MSproxy) registerService(t *Transport) error {
	p := &NetPacket{PacketType: PTypeRegistServer}
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
	PSM.RemServer(t)
	h.reverseRegisterService()
}

func (h *MSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == PTypeRegistServer {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("MSP server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("MSP server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		// 反向注册
		h.reverseRegisterService()
		return
	}
	s := CM.GetClientById(data.From2)
	if s == nil {
		log.Warn("MSproxy client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}

type MCproxy struct {
}

func (h *MCproxy) OnNetMade(t *Transport) {
	log.Debug("MCP-------------s made net")
	CM.AddClient(t)
}

func (h *MCproxy) OnNetLost(t *Transport) {
	log.Debug("MCP-------------s lost net")
	CM.RemoveClient(t)
}

func (h *MCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref("MCP_OnData").Deref()

	if data.PacketType == PTypeRegistServer {
		log.Warn("MCP get register packet", data.Rw.RemoteAddr())
		PSM.GetServerIds(data)
		return
	}
	id := CM.GetClient(data.Rw)
	data.From2 = uint32(id)

	client := PSM.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("MCP get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}
