package p2p

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	//	"net/http/httptest"
	"strconv"
	//	"strings"
)

const KDM_PING uint16 = 654
const KDM_PONG uint16 = 655
const KDM_STORE uint16 = 656
const KDM_FIND_NODE uint16 = 657
const KDM_FIND_NODE_ANSWER uint16 = 658
const KDM_FIND_VALUE uint16 = 659

//const KDM_FIND_VALUE_ANSWER uint16 = 660  //KDM_FIND_VALUE_ANSWER is  same as KDM_FIND_NODE_ANSWER
const KDM_FOUND_VALUE uint16 = 661

const SIZE_OF_IP int = 16
const SIZE_OF_PORT int = 2
const SIZE_OF_ID int = 20
const SIZE_OF_PEER int = SIZE_OF_ID + SIZE_OF_IP + SIZE_OF_PORT
const SIZE_OF_NONCE int = 20
const SIZE_OF_HEADER = 4 + SIZE_OF_PEER + SIZE_OF_NONCE
const SIZE_OF_KEY int = 20

type dhtMessage struct {
	header dhtHeader
	body   dhtBody
	data   []byte
}

type dhtHeader struct {
	size        uint16
	messageType uint16
	senderPeer  peer
	nonce       []byte
}

// TODO: unify "decode" and "parse"
func (h *dhtHeader) decodeHeaderToBytes() []byte {
	result := make([]byte, 0, SIZE_OF_HEADER)
	binary.BigEndian.PutUint16(result[0:2], h.size)
	binary.BigEndian.PutUint16(result[2:4], h.messageType)
	result = append(result, peerToByte(h.senderPeer)...)
	result = append(result, h.nonce...)
	return result
}

type dhtBody interface {
	decodeBodyFromBytes(m *dhtMessage)
	decodeBodyToBytes() []byte
}

type kdmFindValueBody struct {
	id id
}

func (b *kdmFindValueBody) decodeBodyFromBytes(m *dhtMessage) {
	var id [SIZE_OF_ID]byte
	copy(id[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.id = id
}
func (b *kdmFindValueBody) decodeBodyToBytes() []byte {
	return b.id.toByte()
}

type kdmFindNodeBody struct {
	id id
}

func (b *kdmFindNodeBody) decodeBodyFromBytes(m *dhtMessage) {
	var id [SIZE_OF_ID]byte
	copy(id[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.id = id
}
func (b *kdmFindNodeBody) decodeBodyToBytes() []byte {
	return b.id.toByte()
}

type kdmFoundValueBody struct {
	//id id
	value []byte
}

func (b *kdmFoundValueBody) decodeBodyFromBytes(m *dhtMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER
	b.value = m.data[SIZE_OF_HEADER:]
}
func (b *kdmFoundValueBody) decodeBodyToBytes() []byte {
	return b.value
}

type kdmStoreBody struct {
	key   id
	value []byte
}

func (b *kdmStoreBody) decodeBodyFromBytes(m *dhtMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER - SIZE_OF_ID
	var id [SIZE_OF_ID]byte
	copy(id[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.key = id
	b.value = m.data[SIZE_OF_HEADER+SIZE_OF_ID:]
}
func (b *kdmStoreBody) decodeBodyToBytes() []byte {
	var result []byte
	result = append(result, b.key.toByte()...)
	result = append(result, b.value...)
	return result
}

type kdmFindNodeAnswerBody struct {
	answerPeers []peer
}

func (b *kdmFindNodeAnswerBody) decodeBodyFromBytes(m *dhtMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER - SIZE_OF_ID
	var numberOfAnswerPeers = (int(m.header.size) - SIZE_OF_HEADER) / SIZE_OF_PEER
	for i := 0; i < numberOfAnswerPeers; i++ {
		b.answerPeers = append(b.answerPeers, parseByteToPeer(m.data[SIZE_OF_HEADER+i*SIZE_OF_PEER:SIZE_OF_HEADER+(i+1)*SIZE_OF_PEER]))
	}
}
func (b *kdmFindNodeAnswerBody) decodeBodyToBytes() []byte {
	var result []byte
	for i := 0; i < len(b.answerPeers); i++ {
		result = append(result, peerToByte(b.answerPeers[i])...)
	}
	return result
}

func readMessage(conn net.Conn) dhtMessage {
	hdr := dhtHeader{}
	msg := dhtMessage{
		header: hdr,
	}

	//extract size
	messageSize := make([]byte, 2)
	conn.Read(messageSize)
	msg.header.size = binary.BigEndian.Uint16(messageSize)

	//extract all bytes of the message
	messageData := make([]byte, 0, msg.header.size)
	messageData = append(messageData, messageSize...)
	conn.Read(messageData[2:])
	msg.data = messageData

	//extract rest of the header
	msg.header.messageType = binary.BigEndian.Uint16(messageData[2:4])
	msg.header.senderPeer = parseByteToPeer(messageData[4 : 4+SIZE_OF_PEER])
	msg.header.nonce = messageData[4+SIZE_OF_PEER : 4+SIZE_OF_PEER+SIZE_OF_NONCE]

	switch msg.header.messageType {
	case KDM_PING:
	case KDM_PONG:
	case KDM_STORE:
		msg.body = &kdmStoreBody{}
		msg.body.decodeBodyFromBytes(&msg)
	case KDM_FIND_NODE:
		msg.body = &kdmFindNodeBody{}
		msg.body.decodeBodyFromBytes(&msg)
	case KDM_FIND_NODE_ANSWER:
		msg.body = &kdmFindNodeAnswerBody{}
		msg.body.decodeBodyFromBytes(&msg)
	case KDM_FIND_VALUE:
		msg.body = &kdmFindValueBody{}
		msg.body.decodeBodyFromBytes(&msg)
	//case KDM_FIND_VALUE_ANSWER:
	case KDM_FOUND_VALUE:
		msg.body = &kdmFoundValueBody{}
		msg.body.decodeBodyFromBytes(&msg)
	}
	return msg
}

func makeMessage(body dhtBody, msgType uint16) dhtMessage {
	result := dhtMessage{}
	result.header.messageType = msgType
	result.header.senderPeer = n.peer
	nonce := make([]byte, 20)
	if _, err := rand.Read(nonce); err != nil {
		panic(err.Error())
	}
	result.header.nonce = nonce

	if msgType == KDM_PING || msgType == KDM_PONG {
		result.header.size = uint16(SIZE_OF_HEADER)
		result.data = result.header.decodeHeaderToBytes()
	} else {
		bodyData := body.decodeBodyToBytes()
		result.body = body
		result.header.size = uint16(SIZE_OF_HEADER + len(bodyData))
		data := make([]byte, result.header.size)
		data = append(data, result.header.decodeHeaderToBytes()...)
		data = append(data, bodyData...)
	}

	return result
}

//sends the data of message m to the receiver peer
func sendMessage(m dhtMessage, receiverPeer peer) {
	senderTCPaddr, err := net.ResolveTCPAddr("tcp", m.header.senderPeer.ip+":"+strconv.Itoa(int(m.header.senderPeer.port)))
	if err != nil {
		custError := "[FAILURE] Error while parsing to TCP addr: " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	receiverTCPaddr, err := net.ResolveTCPAddr("tcp", receiverPeer.ip+":"+strconv.Itoa(int(receiverPeer.port)))
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
// TODO: peerToByte and parseByteToPeer correct naming
func peerToByte(peer peer) []byte {

	result := make([]byte, 0, SIZE_OF_PEER)
	//first field = ID
	result = append(result, peer.id.toByte()...) //

	//second field = IP
	ip := net.ParseIP(peer.ip).To16()
	result = append(result, ip...)

	//third field = port
	binary.BigEndian.PutUint16(result[(len(peer.id)+len(peer.ip)):], peer.port)
	return result
}

func parseByteToPeer(data []byte) peer {
	var id [SIZE_OF_ID]byte
	copy(id[:], data[:SIZE_OF_ID])

	p := peer{
		id:   id,
		ip:   string(net.IP(data[SIZE_OF_ID : SIZE_OF_ID+SIZE_OF_IP])),
		port: binary.BigEndian.Uint16(data[SIZE_OF_ID+SIZE_OF_IP:]),
	}
	return p
}
