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
	"github.com/dongxiaozhen/lbbref/goref"
	log "github.com/donnie4w/go-logger/logger"
)

type Hello struct {
	lbbnet.NetProcess
}

func (h *Hello) Init() {
	h.NetProcess.Init()
	if stype == 1 {
		h.RegisterFunc(1, fa)
	} else if stype == 2 {
		h.RegisterFunc(2, fb)
	} else if stype == 3 {
		h.RegisterFunc(3, fc)
	} else if stype == 4 {
		h.RegisterFunc(4, fd)
	} else {
		h.RegisterFunc(1, fa)
		h.RegisterFunc(2, fb)
		h.RegisterFunc(3, fc)
		h.RegisterFunc(4, fd)
	}
}

func fa(data *lbbnet.NetPacket) {
	defer goref.Ref("fa").Deref()
	suf := fmt.Sprintf("fa %d,%d,%d,%d,seqid %d", data.UserId, data.From2, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fb(data *lbbnet.NetPacket) {
	defer goref.Ref("fb").Deref()
	log.Debug("fb start")
	suf := fmt.Sprintf("fb %d,%d,%d,%d,seqid %d", data.UserId, data.From2, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fc(data *lbbnet.NetPacket) {
	defer goref.Ref("fc").Deref()
	suf := fmt.Sprintf("fc %d,%d,%d,%d,seqid %d", data.UserId, data.From2, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fd(data *lbbnet.NetPacket) {
	defer goref.Ref("fd").Deref()
	log.Debug("fd start")
	suf := fmt.Sprintf("fd %d,%d,%d,%d,seqid %d", data.UserId, data.From2, data.SessionId, data.PacketType, data.SeqId)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}

var cfg lbbconsul.ConsulConfig
var stype int

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 9627, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "serverNode_2_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "serverNode_2", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:9727", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.IntVar(&stype, "stype", 1, "f1 f2 fall")
	flag.Parse()

	// log.SetLevel(log.WARN)
	log.SetLevel(log.ALL)
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
	hello.Init()
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
