package lbbnet

import (
	"fmt"
	"time"

	"github.com/dongxiaozhen/lbbconsul"

	log "github.com/donnie4w/go-logger/logger"
)

func CompareDiff(old, new map[string]*lbbconsul.ServiceInfo, pf Protocol, pp Manager) {
	for k, v := range old {
		addr := fmt.Sprintf("%s:%d", v.IP, v.Port)
		if v2, ok := new[k]; ok {
			if v2.IP != v.IP || v2.Port != v.Port || (v2.IP == v.IP && v2.Port == v.Port && !pp.HasServer(addr)) {
				log.Warn(" CompareDiff remvoe server------> ", *v)
				pp.RemoveServerByAddr(addr)
				t, err := NewTClient(addr, pf, 0)
				if err != nil {
					log.Warn("CompareDiff proxy server err ", addr, err)
				} else {
					pp.AddTServer(addr, t)
				}
			}
		} else {
			log.Debug("CompareDiff remove server ---> ", *v)
			pp.TmpRemoveServerByAddr(addr)
		}
	}

	for k, v := range new {
		addr := fmt.Sprintf("%s:%d", v.IP, v.Port)
		if _, ok := old[k]; !ok {
			log.Debug("CompareDiff add server ", *v)
			t, err := NewTClient(addr, pf, 0)
			if err != nil {
				log.Warn("CompareDiff proxy client err", addr, err)
			} else {
				pp.AddTServer(addr, t)
			}
		}
	}
}

func MonitorNet(duration int, foundServer string, sp Protocol, pp Manager) {
	tick := time.NewTicker(time.Duration(duration) * time.Second)
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
		CompareDiff(oldSer, services, sp, pp)
		oldSer = services
	}
}
