package lbbnet

import (
	"encoding/json"
	"time"

	"github.com/dongxiaozhen/lbbsort"
	log "github.com/donnie4w/go-logger/logger"
)

type NetProcess struct {
	close      bool
	task       *WorkTask
	mp         map[uint32]func(*NetPacket)
	defun      func(*NetPacket)
	tr         *Transport
	ServerInfo []byte
	ids        map[string]*Transport
	DownQueue  chan *NetPacket
}

func (h *NetProcess) Init() {
	h.task = NewWorkTask(10, 100)
	h.mp = make(map[uint32]func(*NetPacket))
	h.ids = make(map[string]*Transport)
	h.DownQueue = make(chan *NetPacket, 1000)
	h.RegisterServer()
	go h.runDown()
	h.task.Run()
}

func (h *NetProcess) Down(data *NetPacket) {
	log.Warn("lbb down queue", data.PacketType, data.SeqId, data.UserId, data.From1, data.From2)
	h.DownQueue <- data
}

func (h *NetProcess) runDown() {
	for {
		ls := len(h.DownQueue)
		if ls == 0 {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(6 * time.Second)
		for d := range h.DownQueue {
			trp, ok := h.ids[d.Rw.GetRemoteId()]
			if ok {
				log.Warn("lbb down queue second_send ", d.PacketType, d.SeqId, d.UserId, d.From1, d.From2)
				trp.WriteData(d)
			} else {
				log.Error("discard down queue net data ", d.From1, d.From2, d.PacketType, d.Agent, string(d.Data))
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (h *NetProcess) RegisterServer() {
	h.RegisterFunc(PTypeSysObtainServices, h.f)
}

func (h *NetProcess) f(data *NetPacket) {
	s := make([]uint32, 0, len(h.mp))
	for id, _ := range h.mp {
		s = append(s, id)
	}
	lbbsort.Uint32s(s)
	jd, err := json.Marshal(s[1:])
	if err != nil {
		panic("marshal server ids err")
	}
	data.Data = jd
	data.ReqType = MTypeReply
	data.Rw.WriteData(data)
	log.Warn("regist server", string(jd))
}

func (h *NetProcess) RegisterDefFunc(f func(*NetPacket)) error {
	h.defun = f
	return nil
}

func (h *NetProcess) RegisterFunc(packType uint32, f func(*NetPacket)) error {
	if _, ok := h.mp[packType]; ok {
		return ErrFuncFind
	}
	h.mp[packType] = f
	return nil
}

func (h *NetProcess) OnNetMade(t *Transport) {
	log.Warn("NetProcess  made ", t.RemoteAddr())
}

func (h *NetProcess) OnNetLost(t *Transport) {
	log.Warn("NetProcess lost ", t.RemoteAddr())
	sid := t.GetRemoteId()
	delete(h.ids, sid)
}

func (h *NetProcess) OnNetData(data *NetPacket) {
	log.Debug("NetProcess ondata", data.PacketType, data.UserId)
	// if h.close {
	// log.Debug("process close", *data)
	// return
	// }
	if data.PacketType == PTypeSysNotifyServerId && data.ReqType == MTypeOneWay {
		data.Rw.SetRemoteId(string(data.Data))
		h.ids[string(data.Data)] = data.Rw
		log.Warn("add notify server id", string(data.Data))
		return
	}
	if data.ReqType == MTypeRoute {
		data.Data = h.ServerInfo
		data.Rw.WriteData(data)
		return
	}

	hander := h.getHandler(data.PacketType)
	if hander == nil {
		log.Warn("not process this type", data.PacketType)
		return
	}
	h.task.SendTask(func() {
		hander(data)
	})
}

func (h *NetProcess) getHandler(packetType uint32) func(*NetPacket) {
	f, ok := h.mp[packetType]
	if ok {
		return f
	}
	return h.defun
}

func (h *NetProcess) Close() {
	// h.close = true
	h.task.Stop()
	close(h.DownQueue)
}
