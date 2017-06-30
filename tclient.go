package lbbnet

import (
	"errors"
	"net"
	"sync"
	"time"

	log "github.com/donnie4w/go-logger/logger"
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

func (p *TClient) Send(data *NetPacket) error {
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
}

func (p *TClient) connect() error {
	conn, err := net.Dial("tcp", p.addr)
	if err != nil {
		log.Warn("dial err")
		return err
	}
	con, ok := conn.(*net.TCPConn)
	if !ok {
		log.Warn("conver err")
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
		log.Debug("reconnect")
		p.run = false
		p.recon()
	}()

	defer p.pf.OnNetLost(p.transport)
	defer func() {
		log.Debug("transport close")
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
		log.Debug("recon,", err)
	}
}

func (p *TClient) handlerData() {
	defer func() {
		log.Debug("tclient handlerData over")
	}()

	p.run = true
	log.Debug("tclient run true")
	for {
		if s := p.transport.ReadData(); s == nil {
			log.Warn("client handlerData return")
			return
		} else {
			p.pf.OnNetData(s)
		}
	}
}
