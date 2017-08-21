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
	var foundServer string
	flag.StringVar(&foundServer, "fdsvr", "serverNode_2", "found server name")
	flag.Parse()

	log.SetLevel(log.WARN)
	// log.SetLevel(log.ALL)
	log.SetConsole(false)
	log.SetRollingFile("log", "mproxy", 10, 5, log.MB)

	var sproxy = &lbbnet.MSproxy{ServerInfo: lbbconsul.GetConsulInfo()}
	var cproxy = &lbbnet.MCproxy{}

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

	go lbbnet.MonitorNet(2, foundServer, sproxy, lbbnet.PSM)

	lbbutil.RegistSignal(func() {
		lbbconsul.GConsulClient.Close()
		s.Close()
	}, syscall.SIGTERM)
}
