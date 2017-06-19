package lbbnet

import (
	"net"
	"sync"
	"time"

	log "github.com/donnie4w/go-logger/logger"
)

type TServer struct {
	addr    string
	timeout time.Duration
	pf      Protocol
	l       *net.TCPListener
	wg      sync.WaitGroup
}

func NewTServer(addr string, pf Protocol, timeout time.Duration) (*TServer, error) {
	t := &TServer{addr: addr, pf: pf, timeout: timeout}
	err := t.listern()
	return t, err
}

func (p *TServer) Close() {
	p.l.Close()
}

func (p *TServer) listern() error {
	ld, err := net.ResolveTCPAddr("tcp", p.addr)
	if err != nil {
		log.Debug("%#v", err)
		return err
	}
	p.l, err = net.ListenTCP("tcp", ld)
	if err != nil {
		log.Debug("%#v", err)
		return err
	}
	go p.accept()
	return nil
}

func (p *TServer) accept() {
	for {
		con, err := p.l.AcceptTCP()
		if err != nil {
			log.Debug("accept err %#v", err)
			return
		}
		// con.SetReadBuffer(100)
		// con.SetNoDelay(true)
		t := NewTransport(con, p.timeout)
		t.BeginWork()
		p.wg.Add(1)
		go p.handler(t)
	}
}

func (p *TServer) handler(t *Transport) {
	defer p.wg.Done()

	defer t.Close()

	p.pf.OnNetMade(t)
	defer p.pf.OnNetLost(t)

	for {
		log.Debug("ts readdata")
		s := t.ReadData()
		if s == nil {
			log.Debug("ts readdata nil")
			return
		}
		p.pf.OnNetData(s)
	}
}
