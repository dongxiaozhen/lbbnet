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
	var (
		foundServer string
		agentId     int
	)
	flag.StringVar(&foundServer, "fdsvr", "server_proxy", "found server name")
	flag.IntVar(&agentId, "agentId", 1, "agent id")
	flag.Parse()
	sj, err := json.Marshal(lbbconsul.Ccfg)
	if err != nil {
		return
	}

	var sproxy = &lbbnet.FSproxy{ServerInfo: sj}
	var cproxy = &lbbnet.FCproxy{Agent: uint32(agentId), SessionManager: lbbnet.NewSessionManager()}
	log.SetLevel(log.WARN)
	// log.SetLevel(log.ALL)
	log.SetConsole(false)
	log.SetRollingFile("log", "fproxy", 10, 5, log.MB)

	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

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

	go lbbnet.MonitorNet(2, foundServer, sproxy, lbbnet.PSM)

	<-exist
	lbbconsul.GConsulClient.Close()
	s.Close()
}
