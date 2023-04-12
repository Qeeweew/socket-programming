package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"socket-programming/packet"
	"strings"
	"sync"
)

var clientMap = make(map[string]*ClientData)
var passwords = make(map[string]string)

const ip_port = "0.0.0.0:14444"
const BUFSIZE = 1024

type ClientData struct {
	Addr     string       // 网络地址 ip + port
	UserName string       // 用户名
	Conn     *net.TCPConn // TCP connection
	mu       sync.Mutex
}

func (client *ClientData) sendMessage(type_info byte, s string) {
	client.mu.Lock()
	packet.PacketSend(client.Conn, packet.NewPacket(type_info, s))
	client.mu.Unlock()
}

func (client *ClientData) processMessage(p *packet.Packet) {
	msg := string(p.Data)
	switch p.Type {
	case packet.LOGIN:
		i := strings.Index(msg, ",")
		name, password := msg[:i], msg[i+1:]
		password1, isVaild := passwords[name]
		if !isVaild {
			client.sendMessage(packet.FAIL, "No such user")
			break
		}
		if password != password1 {
			client.sendMessage(packet.FAIL, "Wrong password")
			break
		}
		if _, isLogin := clientMap[name]; client.UserName != "" || isLogin {
			client.sendMessage(packet.FAIL, "Already Login")
			break
		}
		client.sendMessage(packet.LOGINSUCCESS, "")
		client.UserName = name
		clientMap[name] = client
	case packet.SEND:
		if client.UserName == "" {
			client.sendMessage(packet.FAIL, "Please login first!")
			break
		}
		i := strings.Index(msg, "$")
		nameTo, msgTo := msg[:i], msg[i+1:]
		clientTo, ok := clientMap[nameTo]
		if ok {
			clientTo.sendMessage(packet.RECEIVE_MESSAGE, fmt.Sprintf("%s$%s", client.UserName, msgTo))
			client.sendMessage(packet.SUCCESS, "")
		} else {
			client.sendMessage(packet.FAIL, fmt.Sprintf("%s is not online", nameTo))
		}
	case packet.SEND_FILE:
		if client.UserName == "" {
			client.sendMessage(packet.FAIL, "Please login first!")
			break
		}
		arr := strings.SplitN(msg, "$", 3)
		nameTo, fileName, msgTo := arr[0], arr[1], arr[2]
		clientTo, ok := clientMap[nameTo]
		if ok {
			// fmt.Println("???")
			clientTo.sendMessage(packet.RECEIVE_FILE, fmt.Sprintf("%s$%s$%s", client.UserName, fileName, msgTo))
			client.sendMessage(packet.SUCCESS, "")
		} else {
			client.sendMessage(packet.FAIL, fmt.Sprintf("%s is not online", nameTo))
		}
	case packet.LOGOUT:
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
		fmt.Printf("%s 已连接\n", conn.RemoteAddr().String())
		go client.receive()
	}
}

func (client *ClientData) receive() {
	for {
		recv_packet, err := packet.PacketReceive(client.Conn)
		if err != nil {
			fmt.Printf("receive err: %v", err)
			break
		}
		client.processMessage(&recv_packet)
		fmt.Println(len(recv_packet.Data))
		if len(recv_packet.Data) < 1000 {
			fmt.Printf("%s -- from: %s\n", string(recv_packet.Data), client.Conn.RemoteAddr().String())
		}
	}
}

func main() {
	file, e := os.Open("passwords.csv")
	if e != nil {
		fmt.Println(e)
		return
	}

	reader := csv.NewReader(file)

	result, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	for _, s := range result {
		passwords[s[0]] = s[1]
	}

	listen()
}
