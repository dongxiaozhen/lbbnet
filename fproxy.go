package lbbnet

import (
	"encoding/json"
	"fmt"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbref/goref"
	"github.com/dongxiaozhen/lbbutil"
	log "github.com/donnie4w/go-logger/logger"
)

type FCproxy struct {
	cm    *ClientManager
	psm   *PServerManager
	Agent uint32
	*SessionManager
}

func NewFCproxy(cm *ClientManager, psm *PServerManager, agent uint32) *FCproxy {
	c := &FCproxy{cm: cm, psm: psm, Agent: agent}
	c.SessionManager = NewSessionManager()
	return c
}

func (h *FCproxy) OnNetMade(t *Transport) {
	log.Debug("FCP-------------s made net")
}

func (h *FCproxy) OnNetLost(t *Transport) {
	log.Debug("FCP-------------s lost net")
	h.cm.RemoveClient(t)
	h.Del(lbbutil.ToUint64(t.GetRemoteId()))
}

func (h *FCproxy) OnNetData(data *NetPacket) {
	defer goref.Ref(lbbconsul.GetConsulServerId()).Deref()
	data.Agent = h.Agent

	if data.PacketType == PTypeSysObtainServices && data.ReqType == MTypeCall {
		log.Warn("FCp remote get Regist-->", data.Rw.RemoteAddr())
		h.psm.GetServerIds(data)
		return
	}

	// login,logout judge
	if h.Exist(data.UserId) {
		if data.PacketType == PTypeLogout {
			// logout
			h.Del(data.UserId)
		}
	} else {
		if data.PacketType == PTypeLogin {
			h.Set(data.UserId, data.Rw)
			data.Rw.SetRemoteId(fmt.Sprintf("%d", data.UserId))
			h.cm.AddClient(data.Rw)
		} else {
			// not login
			return
		}
	}

	id := h.cm.GetClient(data.Rw)
	data.From1 = uint32(id)

	client := h.psm.GetServer(data.UserId, data.PacketType)
	if client == nil {
		log.Warn("FCp get client emtpy", data.PacketType)
		return
	}
	client.WriteData(data)
}

type FSproxy struct {
	cm         *ClientManager
	psm        *PServerManager
	ServerInfo []byte
	ServerId   string
}

func NewFSproxy(cm *ClientManager, psm *PServerManager, serverId string, serverInfo []byte) *FSproxy {
	s := &FSproxy{cm: cm, psm: psm, ServerId: serverId, ServerInfo: serverInfo}
	return s
}

func (h *FSproxy) registerService(t *Transport) error {
	p1 := &NetPacket{PacketType: PTypeSysNotifyServerId, ReqType: MTypeOneWay, Data: []byte(h.ServerId)}
	p := &NetPacket{PacketType: PTypeSysObtainServices, ReqType: MTypeCall}
	t.WriteData(p1)
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
	h.psm.RemServer(t)
}

func (h *FSproxy) OnNetData(data *NetPacket) {
	if data.PacketType == PTypeSysObtainServices && data.ReqType == MTypeReply {
		var ids []uint32
		if err := json.Unmarshal(data.Data, &ids); err != nil {
			log.Error("FSP server id not register", data.Rw.RemoteAddr())
		} else {
			log.Warn("server type", data.Rw.RemoteAddr(), string(data.Data))
		}
		h.psm.AddServer(data.Rw, ids)
		return
	} else if data.PacketType == PTypeSysNotifyServicesChange && data.ReqType == MTypeOneWay {
		err := h.registerService(data.Rw)
		if err != nil {
			log.Error("FSP reverse server regist response  register err", data.Rw.RemoteAddr(), err)
		}
		return
	} else if data.ReqType == MTypeRoute {
		buf := make([]byte, 0, len(h.ServerInfo)+len(data.Data))
		buf = append(buf, h.ServerInfo...)
		buf = append(buf, data.Data...)
		data.Data = buf
	}
	s := h.cm.GetClientById(data.From1)
	if s == nil {
		log.Warn("FSp client off-discard data", data.UserId, data.From1, data.From2, data.SeqId, data.PacketType)
		return
	}
	s.WriteData(data)
}
