package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbnet"
)

type Hello struct {
	lbbnet.NetProcess
}

func (h *Hello) init() {
	h.Init()
	h.RegisterFunc(1, fa)
	h.RegisterFunc(2, fb)
}

func fa(data *lbbnet.NetPacket) {
	fmt.Println("fa start")
	time.Sleep(1 * time.Second)
	suf := fmt.Sprintf("1 %d,%d,%d,%d,seqid %d", data.UserId, data.ServerId, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	fmt.Println(tmp)
	buf := data.Encoder([]byte(tmp))
	data.Rw.WriteData(buf)
}
func fb(data *lbbnet.NetPacket) {
	fmt.Println("fb start")
	time.Sleep(1 * time.Second)
	suf := fmt.Sprintf("2 %d,%d,%d,%d,seqid %d", data.UserId, data.ServerId, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	fmt.Println(tmp)
	buf := data.Encoder([]byte(tmp))
	data.Rw.WriteData(buf)
}

func main() {
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

	hello := &Hello{}
	hello.init()
	s, err := lbbnet.NewTServer(":9099", hello, 2*time.Second)
	if err != nil {
		fmt.Println("create server err", err)
		return
	}

	<-closeChan
	// close listern
	s.Close()
	// not hander new data, wait all done
	hello.Close()
}
