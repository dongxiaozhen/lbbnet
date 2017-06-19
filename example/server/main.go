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

type Hello struct {
	lbbnet.NetProcess
}

func (h *Hello) init() {
	h.Init()
	h.RegisterFunc(1, fa)
	h.RegisterFunc(2, fb)
}

func fa(data *lbbnet.NetPacket) {
	log.Debug("fa start")
	time.Sleep(1 * time.Second)
	suf := fmt.Sprintf("1 %d,%d,%d,%d,seqid %d", data.UserId, data.ServerId, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	buf := data.Encoder([]byte(tmp))
	data.Rw.WriteData(buf)
}
func fb(data *lbbnet.NetPacket) {
	log.Debug("fb start")
	time.Sleep(1 * time.Second)
	suf := fmt.Sprintf("2 %d,%d,%d,%d,seqid %d", data.UserId, data.ServerId, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	buf := data.Encoder([]byte(tmp))
	data.Rw.WriteData(buf)
}

var cfg lbbconsul.ConsulConfig

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 9627, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "serverNode_2_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "serverNode_2", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:9727", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.Parse()

	log.SetLevel(log.WARN)
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		log.Warn("GConsulClient,open err:", err)
		return
	}

	hello := &Hello{}
	hello.init()
	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port), hello, 60*time.Second)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	<-closeChan
	//deregister
	lbbconsul.GConsulClient.Close()
	// close listern
	s.Close()
	// not hander new data, wait all done
	hello.Close()
}
