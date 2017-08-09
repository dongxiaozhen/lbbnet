package lbbnet

import (
	"encoding/binary"
	"errors"

	log "github.com/donnie4w/go-logger/logger"
)

var LenOfNetPacket = 30
var LenOfPacket = LenOfNetPacket + 4

type NetPacket struct {
	UserId     uint64 // 用户ID
	From1      uint32 // 转发使用，标记客户端(第一个代理使用)
	From2      uint32 // 转发使用，标记客户端(第二个代理使用)
	PacketType uint32 // 请求类型
	SessionId  uint32 // 会话ID
	SeqId      uint32 // rpc使用,顺序ID
	ReqType    uint16 // 请求分类 request,response
	Data       []byte // 请求数据
	Rw         *Transport
}

var ErrDataLenLimit = errors.New("data len not enough")

func (t *NetPacket) Decoder(data []byte) error {
	if len(data) < LenOfNetPacket {
		return ErrDataLenLimit
	}
	t.UserId = binary.LittleEndian.Uint64(data[:8])
	t.From1 = binary.LittleEndian.Uint32(data[8:12])
	t.From2 = binary.LittleEndian.Uint32(data[12:16])
	t.PacketType = binary.LittleEndian.Uint32(data[16:20])
	t.SessionId = binary.LittleEndian.Uint32(data[20:24])
	t.SeqId = binary.LittleEndian.Uint32(data[24:28])
	t.ReqType = binary.LittleEndian.Uint16(data[28:30])
	log.Debug("Decoder ", t.UserId, t.From2, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	t.Data = data[LenOfNetPacket:]
	return nil
}

func (t *NetPacket) Encoder(data []byte) []byte {
	if t == nil {
		return nil
	}
	buf := make([]byte, LenOfNetPacket+len(data))
	log.Debug("encoder ", t.UserId, t.From2, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	binary.LittleEndian.PutUint64(buf[:8], t.UserId)
	binary.LittleEndian.PutUint32(buf[8:12], t.From1)
	binary.LittleEndian.PutUint32(buf[12:16], t.From2)
	binary.LittleEndian.PutUint32(buf[16:20], t.PacketType)
	binary.LittleEndian.PutUint32(buf[20:24], t.SessionId)
	binary.LittleEndian.PutUint32(buf[24:28], t.SeqId)
	binary.LittleEndian.PutUint16(buf[28:30], t.ReqType)
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
	log.Debug("encoder ", l, t.UserId, t.From2, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	binary.LittleEndian.PutUint32(buf[:4], uint32(LenOfNetPacket+len(t.Data)))
	binary.LittleEndian.PutUint64(buf[4:12], t.UserId)
	binary.LittleEndian.PutUint32(buf[12:16], t.From1)
	binary.LittleEndian.PutUint32(buf[16:20], t.From2)
	binary.LittleEndian.PutUint32(buf[20:24], t.PacketType)
	binary.LittleEndian.PutUint32(buf[24:28], t.SessionId)
	binary.LittleEndian.PutUint32(buf[28:32], t.SeqId)
	binary.LittleEndian.PutUint16(buf[32:34], t.ReqType)
	copy(buf[LenOfPacket:], t.Data)
	return buf
}
