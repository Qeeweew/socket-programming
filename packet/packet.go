package packet

import (
	"encoding/binary"
	"io"
	"net"
)

const (
	LOGIN = iota
	SEND
	SEND_FILE
	LOGOUT

	FAIL
	RECEIVE_MESSAGE
	RECEIVE_FILE
	LOGINSUCCESS
	SUCCESS
)

type Packet struct {
	Length uint32
	Type   byte   // 包类型
	Data   []byte // 包数据
}

func NewPacket(type_info byte, s string) *Packet {
	data := []byte(s)
	return &Packet{uint32(len(data) + 1), type_info, data}
}

// 封包
func PacketSend(conn net.Conn, packet *Packet) (err error) {
	buf := make([]byte, 5+len(packet.Data))
	binary.BigEndian.PutUint32(buf[0:4], packet.Length)
	buf[4] = packet.Type
	copy(buf[5:], packet.Data)
	// fmt.Printf("send: %s\n", string(buf[4:]))
	_, err = conn.Write(buf)
	return
}

// 拆包
func PacketReceive(conn net.Conn) (packet Packet, err error) {
	lengthBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, lengthBuf)
	//check error
	if err != nil {
		return
	}
	packet.Length = binary.BigEndian.Uint32(lengthBuf)
	bodyBuf := make([]byte, packet.Length)
	_, err = io.ReadFull(conn, bodyBuf)
	//check error
	if err != nil {
		return
	}
	packet.Type = bodyBuf[0]
	packet.Data = bodyBuf[1:]
	// fmt.Printf("receive: %d %s\n", length, string(bodyBuf))
	return
}
