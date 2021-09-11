package main

import "C"
import (
	"crypto/rand"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
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
const SIZE_OF_ID int = 32
const SIZE_OF_PEER int = SIZE_OF_ID + SIZE_OF_IP + SIZE_OF_PORT
const SIZE_OF_NONCE int = 20
const SIZE_OF_HEADER = 4 + SIZE_OF_PEER + SIZE_OF_NONCE

const REPUBLISH_TIME int = 3600 // republish every 3600s

//const SIZE_OF_KEY int = 20     //size of key equals size of id

type p2pMessage struct {
	header p2pHeader
	body   p2pBody
	data   []byte
}

func (m *p2pMessage) toString() string {
	result := "[p2pMessage:] \n"
	result = result + " Header: \n" + m.header.toString()
	result = result + " Body: "
	if m.body == nil {
		result = result + " - "
	} else {
		result = result + m.body.toString()
	}
	result = result + "\n Data: " + bytesToString(m.data)

	return result
}

type p2pHeader struct {
	size        uint16
	messageType uint16
	senderPeer  peer
	nonce       []byte
}

func (h *p2pHeader) decodeHeaderToBytes() []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint16(result[0:2], h.size)
	binary.BigEndian.PutUint16(result[2:4], h.messageType)
	result = append(result, decodePeerToByte(h.senderPeer)...)
	result = append(result, h.nonce...)
	return result
}
func (h *p2pHeader) toString() string {
	result := "      [size = " + strconv.Itoa(int(h.size)) + "]"
	result = result + " [type = " + strconv.Itoa(int(h.messageType)) + "]"
	result = result + "      [senderPeer: " + h.senderPeer.toString() + "] \n"
	//	result = result + "      [nonce: " + bytesToString(h.nonce) + "] \n"
	return result
}

type p2pBody interface {
	decodeBodyFromBytes(m *p2pMessage)
	decodeBodyToBytes() []byte
	toString() string
}

type kdmFindValueBody struct {
	id id
}

func (b *kdmFindValueBody) decodeBodyFromBytes(m *p2pMessage) {
	var id [SIZE_OF_ID]byte
	copy(id[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.id = id
}
func (b *kdmFindValueBody) decodeBodyToBytes() []byte {
	return b.id.toByte()
}
func (b *kdmFindValueBody) toString() string {
	return "[ID: " + bytesToString(b.id.toByte()) + "]"
}

type kdmFindNodeBody struct {
	id id
}

func (b *kdmFindNodeBody) decodeBodyFromBytes(m *p2pMessage) {
	var id [SIZE_OF_ID]byte
	copy(id[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.id = id
}
func (b *kdmFindNodeBody) decodeBodyToBytes() []byte {
	return b.id.toByte()
}
func (b *kdmFindNodeBody) toString() string {
	return "[ID: " + bytesToString(b.id.toByte()) + "]"
}

type kdmFoundValueBody struct {
	key   id
	value []byte
}

func (b *kdmFoundValueBody) decodeBodyFromBytes(m *p2pMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER
	var key id
	copy(key[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	b.key = key
	b.value = m.data[SIZE_OF_HEADER+SIZE_OF_ID:]

}
func (b *kdmFoundValueBody) decodeBodyToBytes() []byte {
	result := b.key.toByte()
	result = append(result, b.value...)
	return result
}
func (b *kdmFoundValueBody) toString() string {
	return "[key: " + bytesToString(b.key.toByte()) + ", value: " + bytesToString(b.value) + "]"
}

type kdmStoreBody struct {
	key   id
	ttl   uint16
	value []byte
}

func (b *kdmStoreBody) decodeBodyFromBytes(m *p2pMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER - SIZE_OF_ID
	idHelp := make([]byte, SIZE_OF_ID)
	copy(idHelp[:], m.data[SIZE_OF_HEADER:SIZE_OF_HEADER+SIZE_OF_ID])
	var i id
	copy(i[:], idHelp)
	b.key = i
	b.ttl = binary.BigEndian.Uint16(m.data[SIZE_OF_HEADER+SIZE_OF_ID : SIZE_OF_HEADER+SIZE_OF_ID+2])
	b.value = m.data[SIZE_OF_HEADER+SIZE_OF_ID+2:]
}
func (b *kdmStoreBody) decodeBodyToBytes() []byte {
	var result []byte
	result = append(result, b.key.toByte()...)

	result = append(result, 0)
	result = append(result, 0)
	binary.BigEndian.PutUint16(result[SIZE_OF_ID:SIZE_OF_ID+2], b.ttl)
	result = append(result, b.value...)
	return result
}
func (b *kdmStoreBody) toString() string {
	return "[Key: " + bytesToString(b.key.toByte()) + "](" + strconv.Itoa(int(b.ttl)) + ")\n     [value:" + bytesToString(b.value) + "]"
}

type kdmFindNodeAnswerBody struct {
	answerPeers []peer
}

func (b *kdmFindNodeAnswerBody) decodeBodyFromBytes(m *p2pMessage) {
	//	valueSize := int(m.header.size)- SIZE_OF_HEADER - SIZE_OF_ID
	var numberOfAnswerPeers = (int(m.header.size) - SIZE_OF_HEADER) / SIZE_OF_PEER
	for i := 0; i < numberOfAnswerPeers; i++ {
		b.answerPeers = append(b.answerPeers, decodeBytesToPeer(m.data[SIZE_OF_HEADER+i*SIZE_OF_PEER:SIZE_OF_HEADER+(i+1)*SIZE_OF_PEER]))
	}
}
func (b *kdmFindNodeAnswerBody) decodeBodyToBytes() []byte {
	var result []byte
	for i := 0; i < len(b.answerPeers); i++ {
		result = append(result, decodePeerToByte(b.answerPeers[i])...)
	}
	return result
}

func (b *kdmFindNodeAnswerBody) toString() string {
	result := "[closestPeers: \n"
	for i := 0; i < len(b.answerPeers); i++ {
		result = result + b.answerPeers[i].toString() + "\n"
	}
	result = result + "]"
	return result
}

func readMessage(conn net.Conn) *p2pMessage {
	receivedMessageRaw := make([]byte, maxMessageLength)
	msgSize, err := conn.Read(receivedMessageRaw)
	log.Debug("received message: ", receivedMessageRaw[:30], " ...")
	if err != nil {
		custError := "[pot. FAILURE] MAIN: Error while reading from connection: " + err.Error() + " (This might be because no more data was sent)"
		log.Error(custError)
		conn.Close()
		return nil
	}
	if msgSize > maxMessageLength {
		custError := "[FAILURE] MAIN: Too much data was sent to us: " + strconv.Itoa(msgSize)
		log.Error(custError)
		conn.Close()
		return nil
	}

	size := binary.BigEndian.Uint16(receivedMessageRaw[:2])
	log.Debug("Received message has size: ", size)
	log.Debug("Received message, data: ", receivedMessageRaw[:size])
	if uint16(msgSize) != size {
		custError := "[FAILURE] MAIN: Message size (" + strconv.Itoa(msgSize) + ") does not match specified 'size': " + strconv.Itoa(int(size))
		log.Error(custError)
		log.Error("!!!", receivedMessageRaw[:msgSize])
		conn.Close()
		return nil
	}
	receivedMsg := makeP2PMessageOutOfBytes(receivedMessageRaw[:msgSize])
	log.Debug("Going to return: ", receivedMsg.toString())
	return &receivedMsg
}

func makeP2PMessageOutOfBytes(messageData []byte) p2pMessage {
	hdr := p2pHeader{}
	msg := p2pMessage{
		header: hdr,
	}

	//extract header
	msg.header.size = binary.BigEndian.Uint16(messageData[:2])
	msg.header.messageType = binary.BigEndian.Uint16(messageData[2:4])
	msg.header.senderPeer = decodeBytesToPeer(messageData[4 : 4+SIZE_OF_PEER])
	msg.header.nonce = messageData[4+SIZE_OF_PEER : 4+SIZE_OF_PEER+SIZE_OF_NONCE]

	//store data in raw
	msg.data = messageData

	//extract body
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

func makeP2PMessageOutOfBody(body p2pBody, msgType uint16) p2pMessage {
	result := p2pMessage{}
	result.header.messageType = msgType

	result.header.senderPeer = thisNode.thisPeer
	nonce := make([]byte, SIZE_OF_NONCE)
	if _, err := rand.Read(nonce); err != nil {
		panic(err.Error())
	}
	result.header.nonce = nonce
	log.Debug(result.header.messageType, result.header.nonce)
	log.Debug(result.header.senderPeer)
	if msgType == KDM_PING || msgType == KDM_PONG {
		result.header.size = uint16(SIZE_OF_HEADER)
		result.data = result.header.decodeHeaderToBytes()
	} else {
		bodyData := body.decodeBodyToBytes()
		result.body = body
		result.header.size = uint16(SIZE_OF_HEADER + len(bodyData))
		var data []byte
		data = append(data, result.header.decodeHeaderToBytes()...)
		data = append(data, bodyData...)
		result.data = data
	}

	return result
}

//sends the data of message m to the receiver peer
func sendP2PMessage(m p2pMessage, receiverPeer peer) {
	_, err := net.ResolveTCPAddr("tcp", m.header.senderPeer.ip+":"+strconv.Itoa(int(m.header.senderPeer.port)))
	if err != nil {
		custError := "[FAILURE] Error while parsing to TCP addr: " + err.Error()
		log.Panic(custError)
	}
	_, err = net.ResolveTCPAddr("tcp", receiverPeer.ip+":"+strconv.Itoa(int(receiverPeer.port)))
	if err != nil {
		custError := "[FAILURE] Error while while parsing to TCP addr:" + err.Error()
		log.Error(custError)
		return
	}
	conn, err := net.Dial("tcp", receiverPeer.ip+":"+strconv.Itoa(int(receiverPeer.port)))
	if err != nil {
		custError := "[FAILURE] Error while connecting via tcp:" + err.Error() + ", for " + receiverPeer.ip + ":" + strconv.Itoa(int(receiverPeer.port))
		log.Error(custError)
		return
	}
	_, err = conn.Write(m.data)
	if err != nil {
		custError := "[FAILURE] Writing to connection failed:" + err.Error()
		log.Error(custError)
		return
	}
}

//parses a peer into byte representation
func decodePeerToByte(peer peer) []byte {

	result := make([]byte, 0, SIZE_OF_PEER)

	//first field = IP
	ip := net.ParseIP(peer.ip).To16()
	result = append(result, ip...)

	//second field = port
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, peer.port)
	result = append(result, port...)
	//third field ID
	result = append(result, peer.id.toByte()...)
	return result

}

func decodeBytesToPeer(data []byte) peer {
	var id [SIZE_OF_ID]byte
	copy(id[:], data[SIZE_OF_IP+SIZE_OF_PORT:])

	p := peer{
		id:   id,
		ip:   net.IP(data[:SIZE_OF_IP]).String(),
		port: binary.BigEndian.Uint16(data[SIZE_OF_IP : SIZE_OF_IP+SIZE_OF_PORT]),
	}
	return p
}

func bytesToString(b []byte) string {
	result := ""
	for i := 0; i < len(b); i++ {
		result = result + strconv.Itoa(int(b[i])) + "-"
	}
	return result
}
