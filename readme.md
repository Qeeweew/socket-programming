# 垃圾 课设
## packet定义
```go
type Packet struct {
	Length uint32 // 包长度
	Data   []byte // 包数据
}
```

## Data部分

client 发出的报文:
1. LOGIN$name 登陆
2. SEND\$name\$msg  发送msg给name用户

server 发出的报文:
1. FAIL$msg   操作失败
2. RECEIVE_MESSAGE\$name\$msg 需要接受来自name的msg