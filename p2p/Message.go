package p2p

import (
	"encoding/binary"
	"net"
)

type message struct {
	sender peer

	// header
	size        uint16
	messageType uint16

	// body
	data []byte
}

func readHeader(conn net.Conn, message *message) {

	//   get message size
	messageSize := make([]byte, 2)
	conn.Read(messageSize)
	message.size = binary.BigEndian.Uint16(messageSize)

	// get message type
	messageType := make([]byte, 2)
	conn.Read(messageType)
	message.messageType = binary.BigEndian.Uint16(messageType)

}
