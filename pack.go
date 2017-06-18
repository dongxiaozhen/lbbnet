package lbbnet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var LenOfNetPacket = 26

type NetPacket struct {
	UserId     uint64 // 用户ID
	ServerId   uint32 // 发送者ID
	PacketType uint32 // 请求类型
	SessionId  uint32 // 会话ID
	SeqId      uint32 // 顺序ID sequence
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
	t.ServerId = binary.LittleEndian.Uint32(data[8:12])
	t.PacketType = binary.LittleEndian.Uint32(data[12:16])
	t.SessionId = binary.LittleEndian.Uint32(data[16:20])
	t.SeqId = binary.LittleEndian.Uint32(data[20:24])
	t.ReqType = binary.LittleEndian.Uint16(data[24:26])
	fmt.Printf("Decoder %d,%d,%d,%d,%d,%d\n", t.UserId, t.ServerId, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	t.Data = data[LenOfNetPacket:]
	return nil
}

func (t *NetPacket) Encoder(data []byte) []byte {
	if t == nil {
		return nil
	}
	buf := make([]byte, LenOfNetPacket+len(data))
	fmt.Printf("encoder %d,%d,%d,%d,%d,%d\n", t.UserId, t.ServerId, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	binary.LittleEndian.PutUint64(buf[:8], t.UserId)
	binary.LittleEndian.PutUint32(buf[8:12], t.ServerId)
	binary.LittleEndian.PutUint32(buf[12:16], t.PacketType)
	binary.LittleEndian.PutUint32(buf[16:20], t.SessionId)
	binary.LittleEndian.PutUint32(buf[20:24], t.SeqId)
	binary.LittleEndian.PutUint16(buf[24:26], t.ReqType)
	copy(buf[LenOfNetPacket:], data)
	return buf
}
func (t *NetPacket) Serialize() []byte {
	if t == nil {
		return nil
	}
	buf := make([]byte, LenOfNetPacket+len(t.Data))
	fmt.Printf("encoder %d,%d,%d,%d,%d,%d\n", t.UserId, t.ServerId, t.SessionId, t.PacketType, t.SeqId, t.ReqType)
	binary.LittleEndian.PutUint64(buf[:8], t.UserId)
	binary.LittleEndian.PutUint32(buf[8:12], t.ServerId)
	binary.LittleEndian.PutUint32(buf[12:16], t.PacketType)
	binary.LittleEndian.PutUint32(buf[16:20], t.SessionId)
	binary.LittleEndian.PutUint32(buf[20:24], t.SeqId)
	binary.LittleEndian.PutUint16(buf[24:26], t.ReqType)
	copy(buf[LenOfNetPacket:], t.Data)
	return buf
}
