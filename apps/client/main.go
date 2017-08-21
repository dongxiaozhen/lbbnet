package main

import (
	"flag"
	"fmt"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbutil"

	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

var user_ids []uint64 = []uint64{121, 122, 124, 123, 125, 126, 128, 127, 129, 130}
var user_len int = 10

type Hello struct {
	close bool
}

func (h *Hello) OnNetMade(t *lbbnet.Transport) {
	log.Warn("connect mad ", t.RemoteAddr())
	go func() {
		i := uint64(0)
		for {
			time.Sleep(500 * time.Millisecond)
			p := &lbbnet.NetPacket{UserId: user_ids[i%10], SessionId: uint32(i), PacketType: uint32(typebase + i%2), Data: []byte(fmt.Sprintf("%s:%d", user_str, i))}
			t.WriteData(p)
			i++
		}
	}()
}

func (h *Hello) OnNetData(data *lbbnet.NetPacket) {
	if h.close {
		log.Warn("OnNetData close")
		return
	}
	log.Warn("recv", string(data.Data))
}
func (h *Hello) OnNetLost(t *lbbnet.Transport) {
	log.Warn("connect lost ", t.RemoteAddr())
}

func (h *Hello) Close() {
	h.close = true
	log.Warn("hello close")
}

var foundServer string
var user_str string
var typebase uint64

func main() {
	flag.StringVar(&foundServer, "fdsvr", "ppproxy", "found server name")
	flag.StringVar(&user_str, "ustr", "hahaha", "say hello")
	flag.Uint64Var(&typebase, "tb", 1, "type base")
	flag.Parse()

	log.SetLevel(log.WARN)
	log.SetConsole(false)
	log.SetRollingFile("log", "client", 10, 5, log.MB)

	exit := lbbutil.MakeSignal(syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	err = lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
	if err != nil {
		log.Warn("discover server err", foundServer)
		return
	}
	services, ok := lbbconsul.GConsulClient.GetAllService(foundServer)
	if !ok {
		log.Warn("not find server err", foundServer)
		return
	}
	hello := &Hello{}
	for _, v := range services {
		_, err := lbbnet.NewTClient(fmt.Sprintf("%s:%d", v.IP, v.Port), hello, 60*time.Second)
		if err != nil {
			log.Debug(err)
			return
		}
	}

	<-exit
	lbbconsul.GConsulClient.Close()
	hello.Close()
	// t.Close()
	time.Sleep(3 * time.Second)
}
