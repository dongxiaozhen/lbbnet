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

var cfg lbbconsul.ConsulConfig
var foundServer string
var sproxy = &lbbnet.PSproxy{}
var cproxy = &lbbnet.Cproxy{}

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 2222, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "proxy_id_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "server_proxy", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:2221", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "serverNode_2", "found server name")
	flag.Parse()
	// log.SetLevel(log.WARN)
	log.SetLevel(log.ALL)
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)
	// SM = newServerManager()

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	s, err := lbbnet.NewTServer(fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port), cproxy, 0)
	if err != nil {
		log.Warn("create server err", err)
		return
	}

	go func() {
		tick := time.NewTicker(2 * time.Second)
		var oldSer = make(map[string]*lbbconsul.ServiceInfo)

		for range tick.C {
			err := lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
			if err != nil {
				log.Warn("discover server err", foundServer)
				continue
			}
			services, ok := lbbconsul.GConsulClient.GetAllService(foundServer)
			if !ok {
				log.Warn("not find server err", foundServer)
			}
			lbbnet.CompareDiff(oldSer, services, sproxy, lbbnet.PSM)
			oldSer = services
		}
	}()

	<-exist
	s.Close()
	lbbconsul.GConsulClient.Close()
}
