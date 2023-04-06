package main

import (
	"fmt"
	"net"
	"socket-programming/packet"
	"strings"
	"sync"
)

var clientMap = make(map[string]*ClientData)

const ip_port = "0.0.0.0:14444"
const BUFSIZE = 1024

/*
client 发出的:
1. LOGIN$name 登陆
2. SEND$name$msg  发送msg给name用户

server 发出的:
1. FAIL$msg   操作失败
2. RECEIVE_MESSAGE$name$msg 需要接受来自name的msg
*/

type ClientData struct {
	Addr     string       // 网络地址 ip + port
	UserName string       // 用户名
	Conn     *net.TCPConn // TCP connection
	mu       sync.Mutex
}

func (client *ClientData) sendMessage(s string) {
	client.mu.Lock()
	packet.PacketSend(client.Conn, packet.NewPacket(s))
	client.mu.Unlock()
}

func (client *ClientData) processMessage(s string) {
	i := strings.Index(s, "$")
	if i == -1 {
		client.sendMessage("FAIL")
		return
	}
	command, msg := s[:i], s[i+1:]
	switch command {
	case "LOGIN":
		name := msg
		if _, isLogin := clientMap[name]; client.UserName != "" || isLogin {
			client.sendMessage("FAIL$已登陆")
			break
		}
		client.sendMessage("LOGINSUCCESS$")
		client.UserName = name
		clientMap[name] = client
	case "SEND":
		if client.UserName == "" {
			client.sendMessage("FAIL$未登录")
			break
		}
		i = strings.Index(msg, "$")
		nameTo, msgTo := msg[:i], msg[i+1:]
		clientTo, ok := clientMap[nameTo]
		if ok {
			clientTo.sendMessage(fmt.Sprintf("RECEIVE_MESSAGE$%s$%s", client.UserName, msgTo))
			client.sendMessage("SUCCESS$")
		} else {
			client.sendMessage(fmt.Sprintf("FAIL$%s is not online", nameTo))
		}
	case "SEND_FILE":
		if client.UserName == "" {
			client.sendMessage("FAIL$please login first")
			break
		}
		arr := strings.SplitN(msg, "$", 3)
		nameTo, fileName, msgTo := arr[0], arr[1], arr[2]
		clientTo, ok := clientMap[nameTo]
		if ok {
			clientTo.sendMessage(fmt.Sprintf("RECEIVE_FILE$%s$%s$%s", client.UserName, fileName, msgTo))
			client.sendMessage("SUCCESS$")
		} else {
			client.sendMessage(fmt.Sprintf("FAIL$%s is not online", nameTo))
		}
	case "LOGOUT":
		client.Conn.Close()
		if client.UserName != "" {
			delete(clientMap, client.UserName)
			break
		}
	}
}

func listen() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ip_port)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Println("开始监听")
	for {
		conn, _ := tcpListener.AcceptTCP()
		client := &ClientData{Conn: conn, mu: sync.Mutex{}}
		clientMap[conn.RemoteAddr().String()] = client
		fmt.Printf("%s 已连接\n", conn.RemoteAddr().String())
		go client.receive()
	}
}

func (client *ClientData) receive() {
	for {
		byteMsg, err := packet.PacketReceive(client.Conn)
		if err != nil {
			fmt.Printf("receive err: %v", err)
			break
		}
		client.processMessage(string(byteMsg))
		fmt.Println(len(byteMsg))
		if len(byteMsg) < 1000 {
			fmt.Printf("%s -- from: %s\n", string(byteMsg), client.Conn.RemoteAddr().String())
		}
	}
}

func main() {
	listen()
}
