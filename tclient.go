package lbbnet

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type TClient struct {
	addr      string
	wg        sync.WaitGroup
	pf        Protocol
	transport *Transport
	close     bool
	run       bool
	timeout   time.Duration
	mu        sync.Mutex
}

func NewTClient(addr string, pf Protocol, timeout time.Duration) (*TClient, error) {
	t := &TClient{addr: addr, pf: pf, timeout: timeout}
	err := t.connect()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (p *TClient) Send(data []byte) error {
	if !p.run {
		return errors.New("tclient close")
	}
	return p.transport.WriteData(data)
}

func (p *TClient) Close() {
	p.mu.Lock()
	p.close = true
	p.mu.Unlock()

	p.transport.Close()
	p.wg.Wait()
}

func (p *TClient) connect() error {
	conn, err := net.Dial("tcp", p.addr)
	if err != nil {
		fmt.Println("dial err")
		return err
	}
	con, ok := conn.(*net.TCPConn)
	if !ok {
		fmt.Println("conver err")
		return errors.New("conv err")
	}

	// 设置成员变量p.transport，不能放在go后边，会导致成员变量为空
	p.transport = NewTransport(con, p.timeout)
	p.transport.BeginWork()
	p.pf.OnNetMade(p.transport)

	go p.handlerConnect()
	return nil
}

func (p *TClient) handlerConnect() {
	defer func() {
		fmt.Println("reconnect")
		p.run = false
		p.recon()
	}()

	p.wg.Add(1)
	defer p.wg.Done()

	defer p.pf.OnNetLost(p.transport)
	defer func() {
		fmt.Println("transport close")
		p.transport.Close()
	}()

	p.handlerData()
}

func (p *TClient) recon() {
	p.mu.Lock()
	if p.close {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	err := p.connect()
	for err != nil {
		time.Sleep(5 * time.Second)
		err = p.connect()
		fmt.Println("recon,", err)
	}
}

func (p *TClient) handlerData() {
	defer func() {
		fmt.Println("tclient handlerData over")
		time.Sleep(2 * time.Second)
	}()

	p.run = true
	fmt.Println("tclient run true")
	for {
		if s := p.transport.ReadData(); s == nil {
			fmt.Println("client handlerData return")
			return
		} else {
			p.pf.OnNetData(s)
		}
	}
}