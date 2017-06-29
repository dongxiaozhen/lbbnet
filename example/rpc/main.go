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
	err := t.Open("127.0.0.1:2222", 0)
	if err != nil {
		fmt.Println("bb", err)
		return
	}

	log.SetLevel(log.WARN)
	time.Sleep(1 * time.Second)
	for i := 0; i < 8; i++ {
		go func() {
			for j := 0; j < 100000; j++ {
				ret, err := t.Call(uint32(j%2+1), *uid, []byte(fmt.Sprintf("l-%d ", j)))
				if err != nil {
					fmt.Println("addd", err)
				} else {
					fmt.Println("aa", ret.UserId, string(ret.Data))
				}
			}
		}()
	}

	// time.Sleep(4 * time.Second)
	// for j := 0; j < 10; j++ {
	// ret, err := t.Call(2, 90002, []byte(fmt.Sprintf("l-%d ", j)))
	// if err != nil {
	// fmt.Println("addd", err)
	// } else {
	// fmt.Println("aa", ret.UserId, string(ret.Data))
	// }
	// }

	time.Sleep(1000 * time.Second)
}
