package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

var foundServer string
var sproxy = &lbbnet.FSproxy{}
var cproxy = &lbbnet.FCproxy{}

func main() {
	flag.StringVar(&foundServer, "fdsvr", "server_proxy", "found server name")
	flag.Parse()

	log.SetLevel(log.WARN)
	// log.SetLevel(log.ALL)
	log.SetConsole(false)
	log.SetRollingFile("log", "fproxy", 10, 5, log.MB)

	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)
	// SM = newServerManager()

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port), cproxy, 0)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	go lbbnet.MonitorNet(2, foundServer, sproxy, lbbnet.PSM)

	<-exist
	lbbconsul.GConsulClient.Close()
	s.Close()
}
