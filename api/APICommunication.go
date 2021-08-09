package api

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

const dhtPUT = 650
const dhtGET = 651
const dhtSUCCESS = 652
const dhtFAILURE = 653
const maxMessageLength = 65535

type DhtAnswer struct {
	success bool
	key     []byte
	value   []byte
}

//listens for TCP connections
func StartAPIDispatcher(apiAddressDHT string) {
	l, err := net.Listen("tcp", apiAddressDHT)
	if err != nil {
		custError := "[FAILURE] Error while listening for connection at" + apiAddressDHT + ": " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	defer l.Close()
	fmt.Println("[SUCCESS] Listening on ", apiAddressDHT)
	for {
		con, err := l.Accept()
		if err != nil {
			custError := "[FAILURE] Error while accepting: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		con.SetDeadline(time.Now().Add(time.Minute * 20)) //Set Timeout

		go handleAPIconnection(con)
	}
}

//listens on one connection for new messages
func handleAPIconnection(con net.Conn) {
	receivedMessage := make([]byte, maxMessageLength)
	msgSize, err := con.Read(receivedMessage)
	if err != nil {
		custError := "[FAILURE] Error while reading from connection: " + err.Error()
		fmt.Println(custError)
		con.Close()
		return
	}
	if msgSize > maxMessageLength {
		custError := "[FAILURE] Too much data was sent to us: " + strconv.Itoa(msgSize)
		fmt.Println(custError)
		con.Close()
		return
	}
	size := binary.BigEndian.Uint16(receivedMessage[:2])
	if uint16(msgSize) != size {
		custError := "[FAILURE] Message size (" + strconv.Itoa(msgSize) + ") does not match specified 'size': " + strconv.Itoa(int(size))
		fmt.Println(custError)
		con.Close()
		return
	}
	msgType := binary.BigEndian.Uint16(receivedMessage[2:4])

	switch msgType {
	case dhtPUT:
		HandlePut(binary.BigEndian.Uint16(receivedMessage[4:6]), receivedMessage[7], receivedMessage[8:20], receivedMessage[20:]) //todo p2p.
	case dhtGET:
		if msgSize != 12 {
			custError := "[FAILURE] Message size (" + strconv.Itoa(msgSize) + ") does not match expected size for a GET message"
			fmt.Println(custError)
			con.Close()
			return
		}
		answer := HandleGet(binary.BigEndian.Uint32(receivedMessage[8:20])) //todo p2p.
		answerMessageSize := 12
		answerMessageType := dhtFAILURE
		if answer.success {
			answerMessageSize = answerMessageSize + len(answer.value)
			answerMessageType = dhtSUCCESS
		}
		answerMessage := make([]byte, 4)
		binary.BigEndian.PutUint16(answerMessage[:2], uint16(answerMessageSize))  //set size
		binary.BigEndian.PutUint16(answerMessage[2:4], uint16(answerMessageType)) //set type
		answerMessage = append(answerMessage, answer.key...)                      //set key
		if answer.success {
			answerMessage = append(answerMessage, answer.value...)
		}
		_, err := con.Write(answerMessage)
		if err != nil {
			custError := "[FAILURE] Error while writing to connection: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		fmt.Println("[SUCCESS] Written answer to connection")

	//todo add customized cases if needed
	default:
		custError := "[FAILURE] Message was of not specified type: " + strconv.Itoa(int(msgType))
		fmt.Println(custError)
		con.Close()
		return
	}

	con.SetDeadline(time.Now().Add(time.Minute * 20)) //Timeout restarted
}

//TODO Replace Dummy in p2p
func HandleGet(key uint32) DhtAnswer {
	return DhtAnswer{}
}

//TODO Replace Dummy in p2p
func HandlePut(ttl uint16, replication byte, key []byte, value []byte) {
}
