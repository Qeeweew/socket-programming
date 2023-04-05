package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var clientMap = make(map[string]*ClientData)

const ip_port = "127.0.0.1:10001"
const BUFSIZE = 1024

type ClientData struct {
	Addr     string       // 网络地址 ip + port
	UserName string       // 用户名
	Conn     *net.TCPConn // TCP connection
	mu       sync.Mutex
}

func (client *ClientData) sendMessage(s string) {
	fmt.Printf("send to %s %s\n", client.UserName, s)
	client.Conn.Write([]byte(s))
}

func (client *ClientData) processMessage(s string) {
	i := strings.Index(s, "$")
	if i == -1 {
		client.sendMessage("FAIL")
		return
	}
	command, msg := s[:i], strings.Trim(s[i+1:], "\n\r ")
	switch command {
	case "LOGIN":
		if client.UserName != "" {
			client.sendMessage("FAIL$已登陆")
			break
		}
		name := msg
		client.UserName = name
		clientMap[name] = client
	case "SEND":
		if client.UserName == "" {
			client.sendMessage("FAIL$未登录")
			break
		}
		i = strings.Index(msg, ":")
		nameTo, msgTo := msg[:i], msg[i+1:]
		clientTo, ok := clientMap[nameTo]
		if ok {
			clientTo.sendMessage(fmt.Sprintf("RECEIVE_MESSAGE$%s:%s", client.UserName, msgTo))
			client.sendMessage("SUCCESS$")
		} else {
			client.sendMessage(fmt.Sprintf("FAIL$%s is not online", nameTo))
		}
	}
}

func listen() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ip_port)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err == nil {
		fmt.Println("端口被占用")
		return
	}
	fmt.Println("开始监听")
	for {
		conn, _ := tcpListener.AcceptTCP()
		client := &ClientData{Conn: conn, mu: sync.Mutex{}}
		clientMap[conn.RemoteAddr().String()] = client
		fmt.Printf("connected with %s\n", conn.RemoteAddr().String())
		go client.receive()
	}
}

func (client *ClientData) receive() {
	defer func() {
		client.Conn.Close()
		if client.UserName != "" {
			delete(clientMap, client.UserName)
		}
	}()
	for {
		byteMsg := make([]byte, BUFSIZE)
		len, err := client.Conn.Read(byteMsg)
		if err != nil {
			break
		}
		go client.processMessage(string(byteMsg[:len]))
		fmt.Printf("%s -- from: %s\n", string(byteMsg[:len]), client.Conn.RemoteAddr().String())
	}
}

func main() {
	listen()
}
