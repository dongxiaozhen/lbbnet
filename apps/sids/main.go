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
	log "github.com/donnie4w/go-logger/logger"
)

var user_id uint64
var user_len int = 10

type Hello struct {
	close bool
}

func (h *Hello) OnNetMade(t *lbbnet.Transport) {
	log.Debug("connect mad")
	p := &lbbnet.NetPacket{UserId: user_id, SessionId: uint32(0), PacketType: uint32(0)}
	t.WriteData(p)
}

func (h *Hello) OnNetData(data *lbbnet.NetPacket) {
	if h.close {
		log.Debug("OnNetData close")
		return
	}
	log.Warn("recv", string(data.Data))
}
func (h *Hello) OnNetLost(t *lbbnet.Transport) {
	log.Debug("connect lost")
}

func (h *Hello) Close() {
	h.close = true
	log.Debug("hello close")
}

var foundServer string

func main() {
	flag.StringVar(&foundServer, "fdsvr", "server_proxy", "found server name")
	flag.Uint64Var(&user_id, "uid", 123, "user_id")
	flag.Parse()

	log.SetLevel(log.WARN)

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

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

	<-closeChan
	lbbconsul.GConsulClient.Close()
	hello.Close()
	// t.Close()
	time.Sleep(3 * time.Second)
}
