package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

func main() {
	var uid = flag.Uint64("uid", 900001, "user id")
	flag.Parse()

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		log.Warn("open return", err)
		return
	}

	t := lbbnet.NewIRpc("user_info", 0)
	err = t.Open()
	if err != nil {
		log.Info("bb", err)
		return
	}

	log.SetLevel(log.ALL)
	// log.SetLevel(log.WARN)
	log.SetConsole(true)
	log.SetRollingFile("log", "proxy", 10, 5, log.MB)

	for j := 0; j < 100000; j++ {
		ret, err := t.Call(uint32(j%2+108), *uid, []byte(fmt.Sprintf("l-%d ", j)))
		if err != nil {
			log.Info("addd", err)
		} else {
			log.Info("aa", ret.UserId, string(ret.Data))
		}
		time.Sleep(1000 * time.Millisecond)
	}

	time.Sleep(1000 * time.Second)
}
