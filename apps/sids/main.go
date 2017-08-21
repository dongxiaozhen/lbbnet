package main

import (
	"flag"
	"fmt"

	"github.com/dongxiaozhen/lbbnet"
	log "github.com/donnie4w/go-logger/logger"
)

func main() {
	flag.Parse()
	t := lbbnet.Rpc{}
	err := t.Open("127.0.0.1:8991", 0)
	if err != nil {
		fmt.Println("bb", err)
		return
	}

	log.SetLevel(log.WARN)
	t.Servers()
}
