package lbbnet

import (
	"errors"
	"fmt"
)

type NetProcess struct {
	close bool
	task  *WorkTask
	mp    map[uint32]func(*NetPacket)
}

func (h *NetProcess) Init() {
	h.task = NewWorkTask(10, 100)
	h.mp = make(map[uint32]func(*NetPacket))
	h.task.Run()
}

var ErrFuncFind = errors.New("函数以注册")

func (h *NetProcess) RegisterFunc(packType uint32, f func(*NetPacket)) error {
	if _, ok := h.mp[packType]; ok {
		return ErrFuncFind
	}
	h.mp[packType] = f
	return nil
}

func (h *NetProcess) OnNetMade(t *Transport) {
	fmt.Println("------------t made")
}
func (h *NetProcess) OnNetLost(t *Transport) {
	fmt.Println("------------t lost")
}
func (h *NetProcess) OnNetData(data *NetPacket) {
	fmt.Println("------------data")
	if h.close {
		fmt.Println("process close", *data)
		return
	}

	hander := h.getHandler(data.PacketType)
	if hander == nil {
		return
	}
	h.task.SendTask(func() {
		hander(data)
	})
}

func (h *NetProcess) getHandler(packetType uint32) func(*NetPacket) {
	return h.mp[packetType]
}

func (h *NetProcess) Close() {
	h.close = true
	h.task.Stop()
}
