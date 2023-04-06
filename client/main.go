package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"socket-programming/packet"
	"strings"
	"sync"
)

const ServerAddr = "47.99.119.54:14444"

var conn *net.TCPConn
var wg sync.WaitGroup
var reader = bufio.NewReader(os.Stdin)

func send() {
	defer wg.Done()
	for {
		fmt.Print("请输入命令: ")
		var cmd string
		fmt.Scanf("%s", &cmd)
		cmd = strings.Trim(cmd, " \n\r")
		switch cmd {
		case "login", "l":
			fmt.Print("请输入用户名: ")
			var name string
			fmt.Scanf("%s", &name)
			err := packet.PacketSend(conn, packet.NewPacket("LOGIN$"+name))
			if err != nil {
				fmt.Printf("err %v", err)
			}
		case "send", "s":
			fmt.Print("请输入好友用户名: ")
			var name string
			fmt.Scanf("%s", &name)
			fmt.Print("请输入发送内容: ")
			var msg = make([]byte, 0)
			for {
				ch, err := reader.ReadByte()
				if ch == '#' || err != nil {
					break
				}
				msg = append(msg, ch)
			}
			err := packet.PacketSend(conn, packet.NewPacket("SEND$"+name+"$"+string(msg)))
			if err != nil {
				fmt.Printf("err %v", err)
			}
		case "quit", "q":
			packet.PacketSend(conn, packet.NewPacket("LOGOUT$"))
			return
		default:
			if len(cmd) > 0 {
				fmt.Printf("未知操作: %s", cmd)
			}
		}
	}
}

func receive() {
	defer wg.Done()
	for {
		buf, err := packet.PacketReceive(conn)
		if err != nil {
			return
		}
		s := string(buf)
		i := strings.Index(s, "$")
		command, msg := s[:i], s[i+1:]
		switch command {
		case "FAIL":
			fmt.Printf("\n\u001b[31merror: %s\u001b[0m\n", msg)
		case "RECEIVE_MESSAGE":
			j := strings.Index(msg, "$")
			fmt.Printf("\n\u001b[32mMessage from %s:\n", msg[0:j])
			fmt.Print(msg[j+1:])
			fmt.Print("\u001b[0m")
		default:
		}
	}
}

func main() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ServerAddr)
	conn_, err := net.DialTCP("tcp", nil, tcpAddr)
	conn = conn_
	if err != nil {
		fmt.Println("connecting to server FAILED!")
		return
	}
	wg.Add(2)
	go send()
	go receive()
	wg.Wait()
}
