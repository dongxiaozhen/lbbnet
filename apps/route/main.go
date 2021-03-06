package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

func main() {
	var uid = flag.Uint64("uid", 900001, "user id")
	flag.Parse()
	t := lbbnet.Rpc{}
	err := t.Open("127.0.0.1:8991", 0)
	if err != nil {
		fmt.Println("bb", err)
		return
	}

	log.SetLevel(log.WARN)
	time.Sleep(1 * time.Second)
	ret, err := t.Login(*uid)
	if err != nil {
		fmt.Println("addd", err)
	} else {
		fmt.Println("aa", ret.UserId, string(ret.Data))
	}

	ret, err = t.Route(uint32(2), *uid)
	if err != nil {
		fmt.Println("addd", err)
	} else {
		fmt.Println("aa", ret.UserId, string(ret.Data))
	}

	ret, err = t.Logout(*uid)
	if err != nil {
		fmt.Println("addd", err)
	} else {
		fmt.Println("aa", ret.UserId, string(ret.Data))
	}
	time.Sleep(1000 * time.Second)
}
