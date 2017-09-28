package lbbnet

import "errors"

// package type  系统消息ID 定义
const (
	PTypeSysObtainServices       uint32 = iota // 向服务请求服务ID
	PTypeSysNotifyServerId                     // 向客户端发送自己的serverID
	PTypeSysNotifyServicesChange               // 向客户端发送服务ID变化请求
	PTypeSysNotifyCloseServer                  // 服务向客户端发送关闭通知
)

// 业务层消息ID定义
const (
	PTypeLogin  uint32 = 10000 + iota // 登录
	PTypeLogout                       // 退出
)

// message type call,reply,oneway
const (
	MTypeCall uint16 = iota + 1
	MTypeReply
	MTypeOneWay
	MTypeRoute
)

// net packet data length limit
var MaxSendPackets uint32 = 100

// err define
var ErrTransportClose = errors.New("链接断开")
var ErrEmptyPacket = errors.New("empty packet")
var ErrMaxPacketLen = errors.New("packet data length overhead")
var ErrDataLenLimit = errors.New("data len not enough")
var ErrRpcTimeOut = errors.New("rpc 请求超时")
var ErrFuncFind = errors.New("函数已经注册")
