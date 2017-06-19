package lbbnet

import (
	"errors"
	"sync"
	"time"

	log "github.com/donnie4w/go-logger/logger"
)

var ErrRpcTimeOut = errors.New("rpc 请求超时")

type RpcRet struct {
	c chan *NetPacket
}

func (p *RpcRet) GetReply() (r *NetPacket, err error) {
	select {
	case r = <-p.c:
	case <-time.After(3 * time.Second):
		err = ErrRpcTimeOut
	}
	return
}

func (p *RpcRet) SetReply(t *NetPacket) {
	p.c <- t
}

type Rpc struct {
	t   *TClient
	mp  map[uint32]*RpcRet
	seq uint32
	sync.Mutex
	timeout time.Duration
}

func (p *Rpc) Open(addr string, timeout time.Duration) error {
	var err error
	p.timeout = timeout
	p.mp = make(map[uint32]*RpcRet)
	p.t, err = NewTClient(addr, p, p.timeout)
	if err != nil {
		return err
	}
	return nil
}

func (p *Rpc) OnNetMade(t *Transport) {
	log.Debug("---------made")
}
func (p *Rpc) OnNetLost(t *Transport) {
	log.Debug("---------lost")
}
func (p *Rpc) OnNetData(t *NetPacket) {
	log.Debug("---------data")
	p.Lock()
	defer p.Unlock()

	ret, ok := p.mp[t.SeqId]
	if ok {
		ret.SetReply(t)
		delete(p.mp, t.SeqId)
		return
	}
	log.Debug("not find seqid", t.SeqId)
}

func (p *Rpc) Call(packType uint32, userId uint64, data []byte) (*NetPacket, error) {
	ret := &RpcRet{c: make(chan *NetPacket, 1)}

	p.Lock()
	p.seq++
	p.mp[p.seq] = ret
	t := &NetPacket{UserId: userId, PacketType: packType, SeqId: p.seq}
	p.Unlock()

	dst := t.Encoder(data)
	err := p.t.Send(dst)
	if err != nil {
		log.Warn("send err", err)
		return nil, err
	}
	pkg, err := ret.GetReply()
	if err != nil {
		p.Lock()
		delete(p.mp, t.SeqId)
		p.Unlock()
		log.Warn("get reply", err)
		return nil, err
	}
	return pkg, nil
}
