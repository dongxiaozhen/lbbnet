package main

import (
	"fmt"
	"time"

	"github.com/dongxiaozhen/lbbnet"
)

func main() {
	t := lbbnet.Rpc{}
	err := t.Open("127.0.0.01:9099")
	if err != nil {
		fmt.Println("bb", err)
		return
	}

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				ret, err := t.Call(1, 90001, []byte(fmt.Sprintf("l-%d ", i)))
				if err != nil {
					fmt.Println("addd", err)
				} else {
					fmt.Println("aa", ret.UserId, string(ret.Data))
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}
	time.Sleep(20 * time.Second)
}
