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
	"fyne.io/fyne/v2/dialog"
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
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter password")

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
	loginButton :=
		widget.NewButton("Login", func() {
			err := packet.PacketSend(conn, packet.NewPacket(packet.LOGIN, loginEntry.Text+","+passwordEntry.Text))
			if err != nil {
				fmt.Printf("err %v", err)
			}
		})
	sendButton :=
		widget.NewButton("Send", func() {
			err := packet.PacketSend(conn, packet.NewPacket(packet.SEND, sendNameEntry.Text+"$"+sendMsgEntry.Text))
			if err != nil {
				fmt.Printf("err %v", err)
			}
		})
	sendFileButton := widget.NewButton("Send a file", func() {
		nameTo := sendNameEntry.Text
		fmt.Printf("sending file to %s\n", nameTo)
		dialogSendWindow := app.NewWindow("")
		b := make([]byte, 1024)
		full := make([]byte, 0)
		send_dialog := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
			defer dialogSendWindow.Close()
			if f == nil {
				return
			}
			for {
				n, err := f.Read(b)
				if err != nil {
					break
				}
				full = append(full, b[:n]...)
			}
			fmt.Printf("%d", len(full))
			packet.PacketSend(conn, packet.NewPacket(packet.SEND_FILE, nameTo+"$"+f.URI().Name()+"$"+string(full)))
			f.Close()
		}, dialogSendWindow)
		send_dialog.Show()
		dialogSendWindow.Resize(fyne.NewSize(600, 400))
		send_dialog.Resize(fyne.NewSize(600, 400))
		dialogSendWindow.SetTitle(fmt.Sprintf("sending file to %s\n", nameTo))
		dialogSendWindow.SetFixedSize(true)
		dialogSendWindow.Show()
	})

	sendNameEntry.Disable()
	sendMsgEntry.Disable()
	sendButton.Disable()
	sendFileButton.Disable()

	content := container.NewVSplit(
		container.NewVBox(
			container.NewHSplit(
				loginEntry,
				passwordEntry,
			),
			loginButton,
			sendNameEntry,
			sendMsgEntry,
			sendButton,
			sendFileButton,
		),
		msgLabel)
	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(300, 600))

	file_receive := func(nameFrom string, filename string, data []byte) {
		fmt.Printf("receiving file from %s\n", nameFrom)
		dialogRecvWindow := app.NewWindow("Receive a file")
		file := dialog.NewFileSave(func(f fyne.URIWriteCloser, e error) {
			if f == nil {
				return
			}
			defer dialogRecvWindow.Close()
			f.Write(data)
			f.Close()
		}, dialogRecvWindow)
		file.Show()
		dialogRecvWindow.Resize(fyne.NewSize(600, 400))
		file.Resize(fyne.NewSize(600, 400))
		dialogRecvWindow.SetTitle(fmt.Sprintf("receiving \"%s\"from %s", filename, nameFrom))
		dialogRecvWindow.SetFixedSize(true)
		dialogRecvWindow.Show()
	}

	receive := func() {
		for {
			p, err := packet.PacketReceive(conn)
			if err != nil {
				return
			}
			msg := string(p.Data)
			switch p.Type {
			case packet.FAIL:
				fmt.Printf("\n\u001b[31merror: %s\u001b[0m\n", msg)
				msgBinding.Append("error: " + msg)
			case packet.RECEIVE_MESSAGE:
				j := strings.Index(msg, "$")
				fmt.Printf("\n\u001b[32mMessage from %s:\n", msg[0:j])
				fmt.Print(msg[j+1:])
				fmt.Print("\u001b[0m")
				msgBinding.Append(msg[0:j] + ": " + msg[j+1:])
			case packet.RECEIVE_FILE:
				arr := strings.SplitN(msg, "$", 3)
				go file_receive(arr[0], arr[1], []byte(arr[2]))
			case packet.LOGINSUCCESS:
				loginEntry.Disable()
				passwordEntry.Disable()
				loginButton.Disable()
				sendNameEntry.Enable()
				sendMsgEntry.Enable()
				sendButton.Enable()
				sendFileButton.Enable()
			default:
			}
		}
	}
	go receive()

	mainWindow.SetOnClosed(func() {
		packet.PacketSend(conn, packet.NewPacket(packet.LOGOUT, ""))
	})
	mainWindow.SetMaster()
	mainWindow.ShowAndRun()
}
