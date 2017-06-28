package main

import (
	"fmt"
	"time"

	"github.com/dongxiaozhen/lbbnet"
)

func main() {
	t := lbbnet.Rpc{}
	err := t.Open("127.0.0.01:9627", 5*time.Second)
	if err != nil {
		fmt.Println("bb", err)
		return
	}

	time.Sleep(1 * time.Second)
	for i := 0; i < 8; i++ {
		go func() {
			for j := 0; j < 100000; j++ {
				ret, err := t.Call(uint32(j%2+1), 90001, []byte(fmt.Sprintf("l-%d ", j)))
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
