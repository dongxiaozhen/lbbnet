package lbbnet

// package  type
const PTypeRegistServer uint32 = 0 // 向服务请求服务ID
const PTypeNotifyServer uint32 = 1 // 向客户端发送自己的serverID

const PTypeLogin uint32 = 10
const PTypeLogout uint32 = 11
const PTypeReverseRegistServer uint32 = 0xFFFFFFFF // 向客户端发送服务ID变化请求

// message type call,reply,oneway
const (
	MTypeCall uint16 = iota + 1
	MTypeReply
	MTypeOneWay
	MTypeRoute
)
