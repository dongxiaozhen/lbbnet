package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbconsul"

	"github.com/dongxiaozhen/lbbnet"
)

type Hello struct {
	close bool
}

func (h *Hello) OnNetMade(t *lbbnet.Transport) {
	fmt.Println("connect mad")
	go func() {
		for i := 0; i < 1000; i++ {
			time.Sleep(1000 * time.Millisecond)
			p := &lbbnet.NetPacket{UserId: user_id, SessionId: uint32(i), PacketType: uint32(1 + i%2)}
			data := p.Encoder([]byte(fmt.Sprintf("%s:%d", user_str, i)))
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

var cfg lbbconsul.ConsulConfig
var foundServer string
var user_id uint64
var user_str string

func main() {
	flag.StringVar(&cfg.ServerId, "sid", "client_id_2", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "client_id", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:9429", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "server_proxy", "found server name")
	flag.StringVar(&user_str, "ustr", "hahaha", "say hello")
	flag.Uint64Var(&user_id, "user_id", 124, "user id")
	flag.Parse()
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		fmt.Println("open return", err)
		return
	}

	err = lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
	if err != nil {
		fmt.Println("discover server err", foundServer)
		return
	}
	services, ok := lbbconsul.GConsulClient.GetAllService(foundServer)
	if !ok {
		fmt.Println("not find server err", foundServer)
		return
	}
	hello := &Hello{}
	for _, v := range services {
		_, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), hello, 60*time.Second)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	<-closeChan
	lbbconsul.GConsulClient.Close()
	hello.Close()
	// t.Close()
	time.Sleep(3 * time.Second)
}
