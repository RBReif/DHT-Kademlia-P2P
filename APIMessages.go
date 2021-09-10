package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const dhtPUT = 650
const dhtGET = 651
const dhtSUCCESS = 652
const dhtFAILURE = 653
const maxMessageLength = 65535

type DhtAnswer struct {
	success bool
	key     id
	value   []byte
}

type apiMessage struct {
	header apiHeader
	body   apiBody
	data   []byte
}

func (m *apiMessage) toString() string {
	result := "[apiMessage:] \n"
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

type apiHeader struct {
	size        uint16
	messageType uint16
}

func (h *apiHeader) decodeHeaderToBytes() []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint16(result[0:2], h.size)
	binary.BigEndian.PutUint16(result[2:4], h.messageType)
	return result
}
func (h *apiHeader) toString() string {
	result := "      [size = " + strconv.Itoa(int(h.size)) + "]"
	result = result + " [type = " + strconv.Itoa(int(h.messageType)) + "] \n"
	return result
}

type apiBody interface {
	decodeBodyFromBytes(m *apiMessage)
	decodeBodyToBytes() []byte
	toString() string
}

type putBody struct {
	ttl         uint16
	replication uint8
	reserved    uint8
	key         id
	value       []byte
}

func (b *putBody) toString() string {
	result := "[ttl: " + strconv.Itoa(int(b.ttl)) + ", replication: " + strconv.Itoa(int(b.replication)) + ", reserved: " + strconv.Itoa(int(b.reserved)) + "\n"
	result = result + "     Key: " + bytesToString(b.key.toByte()) + "]\n,     value:" + bytesToString(b.value) + "]"
	return result
}
func (b *putBody) decodeBodyFromBytes(m *apiMessage) {
	b.ttl = binary.BigEndian.Uint16(m.data[4:6])
	b.replication = m.data[6]
	b.reserved = m.data[7]
	var key [SIZE_OF_ID]byte
	copy(key[:], m.data[8:8+SIZE_OF_ID])
	b.key = key
	b.value = m.data[8+SIZE_OF_ID:]
}
func (b *putBody) decodeBodyToBytes() []byte {
	// implemented for testing purposes and not necessarily needed for the API communication
	result := make([]byte, 4)
	binary.BigEndian.PutUint16(result[0:2], b.ttl)
	result[2] = b.replication
	result[3] = b.reserved
	result = append(result, b.key.toByte()...)
	result = append(result, b.value...)
	return result
}

type getBody struct {
	key id
}

func (b *getBody) toString() string {
	return "[Key: " + bytesToString(b.key.toByte()) + "]"
}
func (b *getBody) decodeBodyFromBytes(m *apiMessage) {
	var key [SIZE_OF_ID]byte
	copy(key[:], m.data[4:4+SIZE_OF_ID])
	b.key = key
}
func (b *getBody) decodeBodyToBytes() []byte {
	return b.key.toByte()
}

type successBody struct {
	key   id
	value []byte
}

func (b *successBody) toString() string {
	return "[Key: " + bytesToString(b.key.toByte()) + "]\n     [value:" + bytesToString(b.value)
}
func (b *successBody) decodeBodyFromBytes(m *apiMessage) {
	//decodeBodyFromBytes of successBody is only needed for testing
	var key [SIZE_OF_ID]byte
	copy(key[:], m.data[4:4+SIZE_OF_ID])
	b.key = key

	b.value = m.data[4+SIZE_OF_ID:]
}
func (b *successBody) decodeBodyToBytes() []byte {
	result := b.key.toByte()
	result = append(result, b.value...)
	return result
}

type failureBody struct {
	key id
}

func (b *failureBody) toString() string {
	return "[Key: " + bytesToString(b.key.toByte()) + "]"
}
func (b *failureBody) decodeBodyFromBytes(m *apiMessage) {
	//decodeBodyFromBytes of failureBody is only needed for testing
	var key [SIZE_OF_ID]byte
	copy(key[:], m.data[4:4+SIZE_OF_ID])
	b.key = key
}
func (b *failureBody) decodeBodyToBytes() []byte {
	return b.key.toByte()
}

/*
makeApiMessageOutOfBytes builds an instance of received bytes of e.g. a dhtGet or a dhtPut message
*/
func makeApiMessageOutOfBytes(messageData []byte) apiMessage {
	//extracting header
	hdr := apiHeader{
		size:        binary.BigEndian.Uint16(messageData[:2]),
		messageType: binary.BigEndian.Uint16(messageData[2:4]),
	}
	msg := apiMessage{
		header: hdr,
	}

	//store data in raw
	msg.data = messageData

	//extract body
	switch msg.header.messageType {
	case dhtPUT:
		msg.body = &putBody{}
		msg.body.decodeBodyFromBytes(&msg)
	case dhtGET:
		msg.body = &getBody{}
		msg.body.decodeBodyFromBytes(&msg)

	//dhtSuccess and dhtFailure are only here for testing purposes
	case dhtSUCCESS:
		msg.body = &successBody{}
		msg.body.decodeBodyFromBytes(&msg)
	case dhtFAILURE:
		msg.body = &failureBody{}
		msg.body.decodeBodyFromBytes(&msg)

	default:
		custError := "[FAILURE] Received Message with unknown Type " + strconv.Itoa(int(msg.header.messageType))
		fmt.Println(custError)
		//panic(custError)
	}

	return msg
}

/*
makeApiMessageOutOfAnswer builds a dhtFailure or a dhtSuccess message out of a received DhtAnswer
*/
func makeApiMessageOutOfAnswer(answer DhtAnswer) apiMessage {
	//building header
	hdr := apiHeader{}
	msg := apiMessage{
		header: hdr,
	}
	if answer.success {
		msg.header.size = uint16(2 + 2 + len(answer.key) + len(answer.value))
		msg.header.messageType = dhtSUCCESS
		msg.body = &successBody{
			key:   answer.key,
			value: answer.value,
		}
	} else {
		msg.header.size = uint16(2 + 2 + len(answer.key))
		msg.header.messageType = dhtFAILURE
		msg.body = &failureBody{
			key: answer.key,
		}
	}
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[:2], msg.header.size)
	binary.BigEndian.PutUint16(data[2:4], msg.header.messageType)
	data = append(data, msg.body.decodeBodyToBytes()...)
	msg.data = data
	return msg
}
