package lbbnet

// package  type
const PTypeRegistServer uint32 = 0
const PTypeReverseRegistServer uint32 = 0xFFFFFFFF

// message type call,reply,oneway
const (
	MTypeCall uint16 = iota + 1
	MTypeReply
	MTypeOneWay
)
