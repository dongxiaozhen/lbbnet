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
* proxy 代理转发
* WorkGroup 实现了线程池

---------------------------------------
### 程序架构

<a href="https://github.com/dongxiaozhen"><img src="https://github.com/dongxiaozhen/lbbnet/blob/master/doc/jiagou.png" width="754" height="470" align="middle"/></a><br/>
图中的箭头方向代表建立连接的方向。
---------------------------------------
* fproxy  代表整个程序的前端反向代理，添加、删除服务不用重启代理
* mproxy  代表可动态增添请求类型代理，添加、删除服务不用重启代理
* proxy   代表固定类型请求代理，添加、删除服务不用重启代理
* server  代表类型处理程序
* client  代表客户端

---------------------------------------

###  服务发现
>*添加新服务
>>	1. 代理根据注册的服务类型发现新服务,并主动连接新服务
>>	2. 代理发送服务注册请求
>>	3. 客户端返回服务注册表(json序列化的数组) 
>>	4. 代理向连接它的客户(可以是fproxy,也可以是client)发送反注册请求
>>	5. 客户发送服务注册请求
>>	6. 代理返回服务注册表 
>
>*断开服务(优雅退出)
>>
>>	1. 代理从注册表里删除该链接注册的服务表，停止转发数据
>>	2. 服务处理完所有缓存的请求后退出，断开连接
>>	3. 代理删除该服务，并向连接它的客户(可以是fproxy,也可以是client)发送反注册请求
>>	4. 客户发送服务注册请求
>>	5. 代理返回服务注册表
