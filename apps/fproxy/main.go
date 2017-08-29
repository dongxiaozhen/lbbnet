package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbnet"
	"github.com/dongxiaozhen/lbbutil"
	log "github.com/donnie4w/go-logger/logger"
)

func main() {
	var (
		foundServer string
		agentId     int
	)
	flag.StringVar(&foundServer, "fdsvr", "server_proxy", "found server name")
	flag.IntVar(&agentId, "agentId", 1, "agent id")
	flag.Parse()

	log.SetLevel(log.WARN)
	// log.SetLevel(log.ALL)
	log.SetConsole(true)
	log.SetRollingFile("log", "fproxy", 10, 5, log.MB)

	cm := lbbnet.NewClientManager()
	psm := lbbnet.NewPServerManager()
	var sproxy = lbbnet.NewFSproxy(cm, psm, lbbconsul.GetConsulServerId(), lbbconsul.GetConsulInfo())
	var cproxy = lbbnet.NewFCproxy(cm, psm, uint32(agentId))
	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port), cproxy, 0)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	err = lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	go lbbnet.MonitorNet(2, foundServer, sproxy, psm)

	lbbutil.RegistSignal(func() {
		lbbconsul.GConsulClient.Close()
		s.Close()
		cm.Free()
	}, syscall.SIGTERM)
}
