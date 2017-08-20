package lbbnet

import (
	"errors"
	"sync"
	"time"

	"github.com/alex023/clock"
	log "github.com/donnie4w/go-logger/logger"
)

// 统一管理定时时间
var gclock = clock.NewClock()

var ErrRpcTimeOut = errors.New("rpc 请求超时")

type RpcRet struct {
	c chan *NetPacket
}

func (p *RpcRet) GetReply() (r *NetPacket, err error) {
	c := make(chan bool, 1)
	job, _ := gclock.AddJobWithInterval(3*time.Second, func() {
		c <- true
	})

	select {
	case r = <-p.c:
		gclock.DelJob(job)
		close(c)
	case <-c:
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
func (p *Rpc) Route(packType uint32, userId uint64) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: packType, ReqType: MTypeRoute}
	return p.call(t)
}

func (p *Rpc) Call(packType uint32, userId uint64, data []byte) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: packType, Data: data}
	return p.call(t)
}

func (p *Rpc) Login(userId uint64) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: PTypeLogin}
	return p.call(t)
}

func (p *Rpc) Logout(userId uint64) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: PTypeLogout}
	return p.call(t)
}

func (p *Rpc) call(t *NetPacket) (*NetPacket, error) {
	ret := &RpcRet{c: make(chan *NetPacket, 1)}

	p.Lock()
	p.seq++
	p.mp[p.seq] = ret
	t.SeqId = p.seq
	p.Unlock()

	log.Debug("--------1---data", *t)
	err := p.t.Send(t)
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
