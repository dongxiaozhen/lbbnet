# lbbnet
tcp server client rpc

##用户需要重写该接口

```go
 type Protocol interface {
     // 新链接创建
      OnNetMade(t *Transport)
      // 处理数据
      OnNetData(data *NetPacket)
      // 链接断开
      OnNetLost(t *Transport)
  }
```
---------------------------------------

*程序优雅退出的策略
	业务层 OnNetData 停止接受新的收据，等老数据都处理完后关闭底层链接

---------------------------------------
* Transport 封装了net.conn,实现了底层的读写
* TClient tcp客户端，实现了断开连接后重连机制
* TServer tcp服务器端
* Rpc 封装了tcpclient，使用seqid，实现了底层的链路共享
* WorkGroup 实现了线程池
