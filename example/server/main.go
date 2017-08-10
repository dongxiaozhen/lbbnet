package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	} else if stype == 5 {
		h.RegisterFunc(5, fe)
	} else if stype == 6 {
		h.RegisterFunc(6, ff)
	} else if stype == 7 {
		h.RegisterFunc(7, fg)
	} else {
		h.RegisterFunc(8, fh)
		h.RegisterFunc(9, fi)
	}
}

func fa(data *lbbnet.NetPacket) {
	defer goref.Ref("fa").Deref()
	suf := fmt.Sprintf("fa %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fb(data *lbbnet.NetPacket) {
	defer goref.Ref("fb").Deref()
	suf := fmt.Sprintf("fb %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fc(data *lbbnet.NetPacket) {
	defer goref.Ref("fc").Deref()
	suf := fmt.Sprintf("fc %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fd(data *lbbnet.NetPacket) {
	defer goref.Ref("fd").Deref()
	suf := fmt.Sprintf("fd %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fe(data *lbbnet.NetPacket) {
	defer goref.Ref("fe").Deref()
	suf := fmt.Sprintf("fe %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func ff(data *lbbnet.NetPacket) {
	defer goref.Ref("ff").Deref()
	suf := fmt.Sprintf("ff %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fg(data *lbbnet.NetPacket) {
	defer goref.Ref("fg").Deref()
	suf := fmt.Sprintf("fg %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fh(data *lbbnet.NetPacket) {
	defer goref.Ref("fh").Deref()
	suf := fmt.Sprintf("fh %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fi(data *lbbnet.NetPacket) {
	defer goref.Ref("fi").Deref()
	suf := fmt.Sprintf("fi %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fj(data *lbbnet.NetPacket) {
	defer goref.Ref("fj").Deref()
	suf := fmt.Sprintf("fj %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, cfg.Port)
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
	// s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port), hello, 60*time.Second)
	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port), hello, 0)
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
