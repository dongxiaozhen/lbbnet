package lbbnet

import (
	"encoding/binary"
	"sync"

	log "github.com/donnie4w/go-logger/logger"
)

var LenOfNetPacket = 42
var LenOfPacket = LenOfNetPacket + 4

var packPoll = sync.Pool{New: func() interface{} { return new(NetPacket) }}

type NetPacket struct {
	UserId     uint64 // 用户ID
	PacketType uint32 // 请求类型
	Agent      uint32 // 入口ID
	Version    uint32 // 版本控制
	SessionId  uint32 // 会话ID
	SeqId      uint32 // rpc使用,顺序ID
	ReqType    uint16 // 请求分类 request,response,onecall
	Data       []byte // 请求数据
	From1      uint32 // 转发使用，跟踪信息路径(第一个代理使用)
	From2      uint32 // 转发使用，跟踪信息路径(第二个代理使用)
	From3      uint32 // 转发使用，跟踪信息路径(第三个代理使用)
	Rw         *Transport
}

func (t *NetPacket) Decoder(data []byte) error {
	if len(data) < LenOfNetPacket {
		return ErrDataLenLimit
	}
	t.UserId = binary.LittleEndian.Uint64(data[:8])
	t.Agent = binary.LittleEndian.Uint32(data[8:12])
	t.Version = binary.LittleEndian.Uint32(data[12:16])
	t.PacketType = binary.LittleEndian.Uint32(data[16:20])
	t.SessionId = binary.LittleEndian.Uint32(data[20:24])
	t.SeqId = binary.LittleEndian.Uint32(data[24:28])
	t.ReqType = binary.LittleEndian.Uint16(data[28:30])
	t.From1 = binary.LittleEndian.Uint32(data[30:34])
	t.From2 = binary.LittleEndian.Uint32(data[34:38])
	t.From3 = binary.LittleEndian.Uint32(data[38:42])
	log.Debug("Decoder ", t.UserId, t.Version, t.SessionId, t.PacketType, t.SeqId, t.ReqType, t.From1, t.From2, t.From3)
	t.Data = data[LenOfNetPacket:]
	return nil
}

func (t *NetPacket) Encoder(data []byte) []byte {
	if t == nil {
		return nil
	}
	buf := make([]byte, LenOfNetPacket+len(data))
	log.Debug("encoder ", t.UserId, t.Version, t.SessionId, t.PacketType, t.SeqId, t.ReqType, t.From1, t.From2, t.From3)
	binary.LittleEndian.PutUint64(buf[:8], t.UserId)
	binary.LittleEndian.PutUint32(buf[8:12], t.Agent)
	binary.LittleEndian.PutUint32(buf[12:16], t.Version)
	binary.LittleEndian.PutUint32(buf[16:20], t.PacketType)
	binary.LittleEndian.PutUint32(buf[20:24], t.SessionId)
	binary.LittleEndian.PutUint32(buf[24:28], t.SeqId)
	binary.LittleEndian.PutUint16(buf[28:30], t.ReqType)
	binary.LittleEndian.PutUint32(buf[30:34], t.From1)
	binary.LittleEndian.PutUint32(buf[34:38], t.From2)
	binary.LittleEndian.PutUint32(buf[38:42], t.From3)
	copy(buf[LenOfNetPacket:], data)
	return buf
}
func (t *NetPacket) Serialize() []byte {
	if t == nil {
		return nil
	}
	log.Debug(" serialize netpacket----------------", *t)
	l := LenOfPacket + len(t.Data)
	buf := make([]byte, l)
	log.Debug("encoder ", l, t.UserId, t.Version, t.SessionId, t.PacketType, t.SeqId, t.ReqType, t.From1, t.From2, t.From3)
	binary.LittleEndian.PutUint32(buf[:4], uint32(LenOfNetPacket+len(t.Data)))
	binary.LittleEndian.PutUint64(buf[4:12], t.UserId)
	binary.LittleEndian.PutUint32(buf[12:16], t.Agent)
	binary.LittleEndian.PutUint32(buf[16:20], t.Version)
	binary.LittleEndian.PutUint32(buf[20:24], t.PacketType)
	binary.LittleEndian.PutUint32(buf[24:28], t.SessionId)
	binary.LittleEndian.PutUint32(buf[28:32], t.SeqId)
	binary.LittleEndian.PutUint16(buf[32:34], t.ReqType)
	binary.LittleEndian.PutUint32(buf[34:38], t.From1)
	binary.LittleEndian.PutUint32(buf[38:42], t.From2)
	binary.LittleEndian.PutUint32(buf[42:46], t.From3)
	copy(buf[LenOfPacket:], t.Data)
	return buf
}

func (t *NetPacket) GetRouteInfo() uint32 {
	if t == nil {
		return 0
	}
	if len(t.Data[LenOfPacket:]) != 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(t.Data[LenOfPacket:])
}

func GetNetPack() *NetPacket {
	return packPoll.Get().(*NetPacket)
}

func (t *NetPacket) Free() {
	packPoll.Put(t)
}
