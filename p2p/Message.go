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
	senderID    id
	responseTo  uint32

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

func makeFIND_NODEanswer(peers []peer) []byte {

	answerMessage := make([]byte, 0, len(peers)*20+24)
	binary.BigEndian.PutUint16(answerMessage[:2], uint16(len(peers)*20+24)) //set size
	binary.BigEndian.PutUint16(answerMessage[2:4], uint16(KDM_FIND_NODE))   //set type
	answerMessage = append(answerMessage, n.id[:]...)
	for i := 0; i < len(peers); i++ {
		answerMessage = append(answerMessage, peers[i].id[:]...)
	}
	return answerMessage
}
