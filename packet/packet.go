package packet

import (
	"encoding/binary"
	"io"
	"net"
)

type Packet struct {
	Length uint32 // 包长度
	Data   []byte // 包数据
}

func NewPacket(s string) *Packet {
	data := []byte(s)
	return &Packet{uint32(len(data)), data}
}

// 封包
func PacketSend(conn net.Conn, packet *Packet) (err error) {
	buf := make([]byte, 4+len(packet.Data))
	binary.BigEndian.PutUint32(buf[0:4], packet.Length)
	copy(buf[4:], packet.Data)
	// fmt.Printf("send: %s\n", string(buf[4:]))
	_, err = conn.Write(buf)
	return
}

// 拆包
func PacketReceive(conn net.Conn) (bodyBuf []byte, err error) {
	lengthBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, lengthBuf)
	if err != nil {
		return
	}
	//check error
	length := binary.BigEndian.Uint32(lengthBuf)
	bodyBuf = make([]byte, length)
	_, err = io.ReadFull(conn, bodyBuf)
	// fmt.Printf("receive: %d %s\n", length, string(bodyBuf))
	//check error
	return
}
