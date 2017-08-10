package lbbnet

import (
	"encoding/json"
	"errors"

	"github.com/dongxiaozhen/lbbsort"
	log "github.com/donnie4w/go-logger/logger"
)

type NetProcess struct {
	close bool
	task  *WorkTask
	mp    map[uint32]func(*NetPacket)
	defun func(*NetPacket)
}

func (h *NetProcess) Init() {
	h.task = NewWorkTask(10, 100)
	h.mp = make(map[uint32]func(*NetPacket))
	h.RegisterServer()
	h.task.Run()
}

var ErrFuncFind = errors.New("函数已经注册")

func (h *NetProcess) RegisterServer() {
	h.RegisterFunc(PTypeRegistServer, h.f)
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
	log.Debug("------------t made")
}
func (h *NetProcess) OnNetLost(t *Transport) {
	log.Debug("------------t lost")
}
func (h *NetProcess) OnNetData(data *NetPacket) {
	log.Debug("NetProcess ondata", data.PacketType, data.UserId)
	if h.close {
		log.Debug("process close", *data)
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
	h.close = true
	h.task.Stop()
}
