package lbbnet

import (
	"fmt"
	"sync"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
	log "github.com/donnie4w/go-logger/logger"
)

type IRpc struct {
	ts         map[string]*TClient
	mp         map[uint32]*RpcRet
	addrToId   map[string]string
	now        *TClient
	nowServer  string
	serverName string
	founding   bool
	seq        uint32
	sync.Mutex
	timeout time.Duration
}

func NewIRpc(rpcName string, timeout time.Duration) *IRpc {
	rpc := &IRpc{}
	rpc.serverName = rpcName
	rpc.timeout = timeout
	rpc.mp = make(map[uint32]*RpcRet)
	rpc.ts = make(map[string]*TClient)
	rpc.addrToId = make(map[string]string)
	return rpc
}

func (p *IRpc) Open() error {
	svr, err := FoundRpc(p.serverName, p.nowServer)
	if err != nil {
		log.Error("open iprc err: ", err)
		return err
	}
	addr := fmt.Sprintf("%s:%d", svr.IP, svr.Port)
	t, err := NewTClient(addr, p, p.timeout)
	if err != nil {
		log.Error("rpc dial err:not find server", err)
		return err
	}
	p.Lock()
	p.ts[addr] = t
	p.now = t
	p.nowServer = svr.ServiceID
	p.addrToId[addr] = svr.ServiceID
	p.Unlock()
	log.Warn("-------------iprc get server found ", addr)

	return nil
}

func (p *IRpc) OnNetMade(t *Transport) {
	log.Debug("---------made")
	p1 := &NetPacket{PacketType: PTypeSysNotifyServer, ReqType: MTypeOneWay, Data: []byte(lbbconsul.Ccfg.ServerId)}
	t.WriteData(p1)
}

func (p *IRpc) OnNetLost(t *Transport) {
	log.Debug("---------lost ", t.RemoteAddr())
	p.Lock()
	defer p.Unlock()

	addr := t.RemoteAddr()
	sid := p.addrToId[addr]
	p.ts[addr].Close()
	delete(p.addrToId, addr)
	delete(p.ts, addr)

	if sid == p.nowServer {
		go p.getServer()
	}
}

func (p *IRpc) getServer() {
	log.Debug("----------irpc get server")
	p.Lock()
	if p.founding == false {
		p.founding = true
	} else {
		log.Debug("----------irpc get server, already have getServer run")
		p.Unlock()
		return
	}
	p.Unlock()

	tick := time.NewTicker(time.Second)
	defer func() {
		tick.Stop()
		p.Lock()
		p.founding = false
		p.Unlock()
	}()

	var err error
	for range tick.C {
		if err = p.Open(); err != nil {
			continue
		}
		return
	}
}
func (p *IRpc) OnNetData(t *NetPacket) {
	log.Debug("---------data")
	p.Lock()
	defer p.Unlock()

	if t.PacketType == PTypeSysCloseServer && t.ReqType == MTypeOneWay {
		log.Debug("get PTypeSysCloseServer ", t.Rw.RemoteAddr())
		go p.getServer()
		return
	}

	ret, ok := p.mp[t.SeqId]
	if ok {
		ret.SetReply(t)
		delete(p.mp, t.SeqId)
		return
	}
	log.Debug("not find seqid", t.SeqId)
}

func (p *IRpc) Route(packType uint32, userId uint64) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: packType, ReqType: MTypeRoute}
	return p.call(t)
}

func (p *IRpc) Call(packType uint32, userId uint64, data []byte) (*NetPacket, error) {
	t := &NetPacket{UserId: userId, PacketType: packType, Data: data}
	return p.call(t)
}

func (p *IRpc) Servers() (*NetPacket, error) {
	t := &NetPacket{PacketType: PTypeSysRegistServer, ReqType: MTypeCall}
	return p.call(t)
}

func (p *IRpc) call(t *NetPacket) (*NetPacket, error) {
	ret := &RpcRet{c: make(chan *NetPacket, 1)}

	p.Lock()
	p.seq++
	p.mp[p.seq] = ret
	t.SeqId = p.seq
	p.Unlock()

	err := p.now.Send(t)
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
