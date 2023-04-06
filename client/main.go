package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"socket-programming/packet"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const ServerAddr = "47.99.119.54:14444"

var conn *net.TCPConn
var wg sync.WaitGroup
var reader = bufio.NewReader(os.Stdin)
var receivedMessages []string

func main() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ServerAddr)
	conn_, err := net.DialTCP("tcp", nil, tcpAddr)
	conn = conn_
	if err != nil {
		fmt.Println("connecting to server FAILED!")
		return
	}
	app := app.New()

	mainWindow := app.NewWindow("client")
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Enter user name...")
	sendNameEntry := widget.NewEntry()
	sendNameEntry.SetPlaceHolder("To Whom...")
	sendMsgEntry := widget.NewMultiLineEntry()
	sendMsgEntry.SetPlaceHolder("Message...")
	msgBinding := binding.NewStringList()
	msgLabel := widget.NewListWithData(msgBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	content := container.NewVSplit(
		container.NewVBox(
			container.NewHSplit(
				loginEntry,
				widget.NewButton("Login", func() {
					err := packet.PacketSend(conn, packet.NewPacket("LOGIN$"+loginEntry.Text))
					if err != nil {
						fmt.Printf("err %v", err)
					}
				})),
			sendNameEntry,
			sendMsgEntry,
			widget.NewButton("Send", func() {
				err := packet.PacketSend(conn, packet.NewPacket("SEND$"+sendNameEntry.Text+"$"+sendMsgEntry.Text))
				if err != nil {
					fmt.Printf("err %v", err)
				}
			})),
		msgLabel,
	)
	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(200, 400))

	receive := func() {
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
				newWindow := app.NewWindow("Error")
				label := widget.NewLabel(msg)
				newWindow.SetContent(container.NewCenter(label))
				fmt.Printf("\n\u001b[31merror: %s\u001b[0m\n", msg)
				msgBinding.Append("error: " + msg)
			case "RECEIVE_MESSAGE":
				j := strings.Index(msg, "$")
				newWindow := app.NewWindow(fmt.Sprintf("Received a message from %s", msg[0:j]))
				label := widget.NewLabel(msg[j+1:])
				newWindow.SetContent(container.NewCenter(label))
				fmt.Printf("\n\u001b[32mMessage from %s:\n", msg[0:j])
				fmt.Print(msg[j+1:])
				fmt.Print("\u001b[0m")
				msgBinding.Append(msg[0:j] + ": " + msg[j+1:])
			case "LOGINSUCCESS":
				loginEntry.Disable()
			default:
			}
		}
	}
	go receive()
	mainWindow.SetOnClosed(func() {
		packet.PacketSend(conn, packet.NewPacket("LOGOUT$"))
	})
	mainWindow.ShowAndRun()
}
