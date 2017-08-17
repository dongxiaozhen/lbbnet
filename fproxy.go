package lbbnet

import (
	"encoding/json"

	log "github.com/donnie4w/go-logger/logger"
	"github.com/mreithub/goref"
)

type FCproxy struct {
	Agent uint32
}

func (h *FCproxy) OnNetMade(t *Transport) {
	log.Debug("FCP-------------s made net")
	CM.AddClient(t)
}

func (h *FCproxy) OnNetLost(t *Transport) {
	log.Debug("FCP-------------s lost net")
	CM.RemoveClient(t)
}

func (h *FCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref("FCproxy_OnData").Deref()
	data.Agent = h.Agent

	if data.PacketType == PTypeRegistServer && data.ReqType == MTypeCall {
		log.Warn("FCp remote get Regist-->", data.Rw.RemoteAddr())
		PSM.GetServerIds(data)
		return
	}
	id := CM.GetClient(data.Rw)
	data.From1 = uint32(id)

	client := PSM.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("FCp get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}

type FSproxy struct {
}

func (h *FSproxy) registerService(t *Transport) error {
	p := &NetPacket{PacketType: PTypeRegistServer, ReqType: MTypeCall}
	return t.WriteData(p)
}
func (h *FSproxy) OnNetMade(t *Transport) {
	log.Warn("FSP--------->made", t.RemoteAddr())
	err := h.registerService(t)
	if err != nil {
		log.Error("FSP send server register", t.RemoteAddr())
	}
}

func (h *FSproxy) OnNetLost(t *Transport) {
	log.Warn("FSP--------->lost", t.RemoteAddr())
	PSM.RemServer(t)
}

func (h *FSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == PTypeRegistServer && data.ReqType == MTypeReply {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("FSP server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		PSM.AddServer(data.Rw, ids)
		return
	} else if data.PacketType == PTypeReverseRegistServer && data.ReqType == MTypeOneWay {
		err := h.registerService(data.Rw)
		if err != nil {
			log.Error("FSP reverse server regist response  register err", data.Rw.RemoteAddr(), err)
		}
		return
	}
	s := CM.GetClientById(data.From1)
	if s == nil {
		log.Warn("FSp client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}
