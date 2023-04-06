package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const ServerAddr = "47.99.119.54:14444"

var loginName string
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
			_, err := conn.Write([]byte("LOGIN$" + name))
			if err != nil {
				break
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
			_, err := conn.Write(append([]byte("SEND$"+name+":"), msg...))
			if err != nil {
				break
			}
		case "quit", "q":
			conn.Write(append([]byte("LOGOUT$")))
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
	buf := make([]byte, 1024)
	for {
		len, err := conn.Read(buf)
		if err != nil {
			return
		}
		s := string(buf[:len])
		i := strings.Index(s, "$")
		command, msg := s[:i], s[i+1:]
		switch command {
		case "FAIL":
			fmt.Printf("\n\u001b[31merror: %s\u001b[0m\n", msg)
		case "RECEIVE_MESSAGE":
			j := strings.Index(msg, ":")
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
