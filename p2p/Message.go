package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type message struct {
	sender   peer
	receiver peer

	// header
	size        uint16
	messageType uint16
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

//sends the data of message m to the receiver in message m
func sendMessage(m message) {
	senderTCPaddr, err := net.ResolveTCPAddr("tcp", m.sender.ip+":"+strconv.Itoa(int(m.sender.port)))
	if err != nil {
		custError := "[FAILURE] Error while parsing to TCP addr: " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	receiverTCPaddr, err := net.ResolveTCPAddr("tcp", m.receiver.ip+":"+strconv.Itoa(int(m.sender.port)))
	if err != nil {
		custError := "[FAILURE] Error while while parsing to TCP addr:" + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	conn, err := net.DialTCP("tcp", senderTCPaddr, receiverTCPaddr)
	if err != nil {
		custError := "[FAILURE] Error while connecting via tcp:" + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	_, err = conn.Write(m.data)
	if err != nil {
		custError := "[FAILURE] Writing to connection failed:" + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
}

//parses a peer into byte representation
func peerToByte(peer peer) []byte {

	result := make([]byte, 0, SIZE_OF_ID+SIZE_OF_IP+SIZE_OF_PORT)
	//first field = ID
	result = append(result, peer.id[:]...) //

	//second field = IP
	if strings.Contains(peer.ip, ".") { //ipv4
		var temp [12]byte
		result = append(result, temp[:]...)
	}
	ip := net.ParseIP(peer.ip)
	result = append(result, ip[:]...)

	//third field = port
	binary.BigEndian.PutUint32(result[(len(peer.id)+len(peer.ip)):], peer.port)

	return result
}

// makes a FIND_NODE_ANSWER message including the specified peers
func makeFIND_NODE_ANSWERmessage(peers []peer) []byte {

	size := (len(peers)+1)*40 + 4 //

	answerMessage := make([]byte, 0, size)

	binary.BigEndian.PutUint16(answerMessage[:2], uint16(size))                  //set size
	binary.BigEndian.PutUint16(answerMessage[2:4], uint16(KDM_FIND_NODE_ANSWER)) //set type

	localPeerByte := peerToByte(n.peer)
	answerMessage = append(answerMessage, localPeerByte...) //set the senderPeer

	for i := 0; i < len(peers); i++ {
		answerMessage = append(answerMessage, peerToByte(peers[i])...) //set the peers that are sent as answer
	}

	return answerMessage
}

//makes a FIND_NODE message for a specified key
func makeFIND_NODEmessage(key [20]byte) []byte {

	size := 4 + SIZE_OF_ID + SIZE_OF_IP + SIZE_OF_PORT + SIZE_OF_KEY //
	answerMessage := make([]byte, 0, size)

	binary.BigEndian.PutUint16(answerMessage[:2], uint16(size))           //set size
	binary.BigEndian.PutUint16(answerMessage[2:4], uint16(KDM_FIND_NODE)) //set type

	localPeerByte := peerToByte(n.peer)
	answerMessage = append(answerMessage, localPeerByte...) //set the senderPeer
	answerMessage = append(answerMessage, key[:]...)
	return answerMessage
}

//returns/extracts the peers that were received in a FIND_NODE_ANSWER message
func parseFIND_NODE_ANSWER(m message) []peer {
	var result []peer
	numberOfPeers := int((m.size - 44) / 40)
	for i := 0; i < numberOfPeers; i++ {
		var id [20]byte
		copy(id[:], m.data[:20])

		p := peer{
			id:   id,
			ip:   string(m.data[20:36]),
			port: binary.BigEndian.Uint32(m.data[36:]),
		}
		result = append(result, p)
	}
	return result
}
