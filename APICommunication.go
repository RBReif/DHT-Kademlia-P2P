package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

//listens for TCP connections
func startAPIDispatcher(apiAddressDHT string) {
	l, err := net.Listen("tcp", apiAddressDHT+":"+strconv.Itoa(int(Conf.apiPort)))
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
	receivedMessageRaw := make([]byte, maxMessageLength)
	msgSize, err := con.Read(receivedMessageRaw)
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
	size := binary.BigEndian.Uint16(receivedMessageRaw[:2])
	if uint16(msgSize) != size {
		custError := "[FAILURE] Message size (" + strconv.Itoa(msgSize) + ") does not match specified 'size': " + strconv.Itoa(int(size))
		fmt.Println(custError)
		con.Close()
		return
	}
	receivedMsg := makeApiMessageOutOfBytes(receivedMessageRaw[:msgSize])

	switch receivedMsg.header.messageType {
	case dhtPUT:
		handlePut(receivedMsg.body.(*putBody))

	case dhtGET:
		if receivedMsg.header.size != 12 {
			custError := "[FAILURE] Message size (" + strconv.Itoa(msgSize) + ") does not match expected size for a GET message"
			fmt.Println(custError)
			con.Close()
			return
		}
		answer := handleGet(receivedMsg.body.(*getBody))
		answerMessage := makeApiMessageOutOfAnswer(answer)

		_, err := con.Write(answerMessage.data)
		if err != nil {
			custError := "[FAILURE] Error while writing to connection: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		fmt.Println("[SUCCESS] Written answer to connection")

	//todo add customized cases if needed
	default:
		custError := "[FAILURE] Message was of not specified type: " + strconv.Itoa(int(receivedMsg.header.messageType))
		fmt.Println(custError)
		con.Close()
		return
	}

	con.SetDeadline(time.Now().Add(time.Minute * 20)) //Timeout restarted
}

//TODO Replace Dummy in p2p
func handleGet(body *getBody) DhtAnswer {
	return DhtAnswer{}
}

//TODO Replace Dummy in p2p
func handlePut(body *putBody) {
}
