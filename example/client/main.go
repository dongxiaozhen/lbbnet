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
	close bool
}

func (h *Hello) OnNetMade(t *lbbnet.Transport) {
	fmt.Println("connect mad")
	go func() {
		for i := 0; i < 1000; i++ {
			time.Sleep(100 * time.Millisecond)
			p := &lbbnet.NetPacket{UserId: 123, ServerId: 1, SessionId: 2, PacketType: uint32(1 + i%2)}
			data := p.Encoder([]byte(fmt.Sprintf("client%d", i)))
			t.WriteData(data)
		}
	}()
}

func (h *Hello) OnNetData(data *lbbnet.NetPacket) {
	if h.close {
		fmt.Println("OnNetData close")
		return
	}
	time.Sleep(1 * time.Second)
	fmt.Println("recv", string(data.Data))
}
func (h *Hello) OnNetLost(t *lbbnet.Transport) {
	fmt.Println("connect lost")
}

func (h *Hello) Close() {
	h.close = true
	fmt.Println("hello close")
}

func main() {
	fmt.Println("vim-go")
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

	hello := &Hello{}
	t, err := lbbnet.NewTClient("127.0.0.01:9099", hello)
	if err != nil {
		fmt.Println(err)
		return
	}

	<-closeChan
	hello.Close()
	t.Close()
	time.Sleep(3 * time.Second)
}
