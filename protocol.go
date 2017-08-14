package lbbnet

type Protocol interface {
	OnNetMade(t *Transport)
	OnNetData(data *NetPacket)
	OnNetLost(t *Transport)
}

type Manager interface {
	RemoveServerByAddr(string)
	AddTServer(string, *TClient)
	TmpRemoveServerByAddr(string)
	HasServer(addr string) bool
}
