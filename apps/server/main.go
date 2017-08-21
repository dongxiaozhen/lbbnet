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
	} else if stype == 10 {
		h.RegisterFunc(10, login)
		h.RegisterFunc(11, logout)
	} else {
		h.RegisterFunc(8, fh)
		h.RegisterFunc(9, fi)
	}
}
func login(data *lbbnet.NetPacket) {
	data.Data = []byte("login success")
	data.Rw.WriteData(data)
}
func logout(data *lbbnet.NetPacket) {
	data.Data = []byte("logout success")
	data.Rw.WriteData(data)
}

func fa(data *lbbnet.NetPacket) {
	defer goref.Ref("fa").Deref()
	suf := fmt.Sprintf("fa %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fb(data *lbbnet.NetPacket) {
	defer goref.Ref("fb").Deref()
	suf := fmt.Sprintf("fb %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fc(data *lbbnet.NetPacket) {
	defer goref.Ref("fc").Deref()
	suf := fmt.Sprintf("fc %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fd(data *lbbnet.NetPacket) {
	defer goref.Ref("fd").Deref()
	suf := fmt.Sprintf("fd %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fe(data *lbbnet.NetPacket) {
	defer goref.Ref("fe").Deref()
	suf := fmt.Sprintf("fe %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func ff(data *lbbnet.NetPacket) {
	defer goref.Ref("ff").Deref()
	suf := fmt.Sprintf("ff %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fg(data *lbbnet.NetPacket) {
	defer goref.Ref("fg").Deref()
	suf := fmt.Sprintf("fg %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fh(data *lbbnet.NetPacket) {
	defer goref.Ref("fh").Deref()
	suf := fmt.Sprintf("fh %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fi(data *lbbnet.NetPacket) {
	defer goref.Ref("fi").Deref()
	suf := fmt.Sprintf("fi %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}
func fj(data *lbbnet.NetPacket) {
	defer goref.Ref("fj").Deref()
	suf := fmt.Sprintf("fj %d,%d,%d,%d,seqid %d,%d  <--", data.PacketType, data.UserId, data.From1, data.From2, data.SeqId, lbbconsul.Ccfg.Port)
	tmp := suf + string(data.Data)
	log.Debug(tmp)
	data.Data = []byte(tmp)
	data.Rw.WriteData(data)
}

var stype int

func main() {
	flag.IntVar(&stype, "stype", 1, "f1 f2 fall")
	flag.Parse()

	log.SetLevel(log.WARN)
	log.SetConsole(false)
	// log.SetLevel(log.ALL)
	log.SetRollingFile("log", "server_log", 10, 5, log.MB)

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM)

	hello := &Hello{NetProcess: lbbnet.NetProcess{ServerInfo: lbbconsul.GetConsulInfo()}}
	hello.Init()
	// s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port), hello, 60*time.Second)
	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port), hello, 0)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	err = lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("GConsulClient,open err:", err)
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
