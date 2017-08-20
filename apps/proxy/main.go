package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

func main() {
	var foundServer string
	flag.StringVar(&foundServer, "fdsvr", "serverNode_2", "found server name")
	flag.Parse()

	log.SetLevel(log.WARN)
	log.SetConsole(false)
	// log.SetLevel(log.ALL)
	log.SetRollingFile("log", "proxy", 10, 5, log.MB)

	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	sj, err := json.Marshal(lbbconsul.Ccfg)
	if err != nil {
		return
	}
	var sproxy = &lbbnet.Sproxy{ServerInfo: sj}
	var cproxy = &lbbnet.Cproxy{}

	err = lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port), cproxy, 0)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	go lbbnet.MonitorNet(2, foundServer, sproxy, lbbnet.SM)

	<-exist
	s.Close()
	lbbconsul.GConsulClient.Close()
}
