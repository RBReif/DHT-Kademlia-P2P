package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

//listens for TCP connections
func startAPIDispatcher() {
	l, err := net.Listen("tcp", Conf.apiIP+":"+strconv.Itoa(int(Conf.apiPort)))
	if err != nil {
		custError := "[FAILURE] MAIN: Error while listening for connection at" + Conf.apiIP + ": " + strconv.Itoa(int(Conf.apiPort)) + " - " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	defer l.Close()
	fmt.Println("[SUCCESS] MAIN: Listening on ", Conf.apiIP, ": ", Conf.apiPort)
	for {
		con, err := l.Accept()
		if err != nil {
			custError := "[FAILURE] MAIN: Error while accepting: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		fmt.Println("[SUCCESS] MAIN: New Connection established, ", con)
		con.SetDeadline(time.Now().Add(time.Minute * 20)) //Set Timeout

		go handleAPIconnection(con)
	}
}

//listens on one connection for new messages
func handleAPIconnection(con net.Conn) {
	for {
		receivedMessageRaw := make([]byte, maxMessageLength)
		msgSize, err := con.Read(receivedMessageRaw)
		//	fmt.Println("received message: ", receivedMessageRaw[:30], " ...")
		if err != nil {
			custError := "[FAILURE] MAIN: Error while reading from connection: " + err.Error()
			fmt.Println(custError)
			con.Close()
			return
		}
		if msgSize > maxMessageLength {
			custError := "[FAILURE] MAIN: Too much data was sent to us: " + strconv.Itoa(msgSize)
			fmt.Println(custError)
			con.Close()
			return
		}
		size := binary.BigEndian.Uint16(receivedMessageRaw[:2])
		fmt.Println("Received message has size: ", size)
		fmt.Println("Received message, data: ", receivedMessageRaw[:size])
		if uint16(msgSize) != size {
			custError := "[FAILURE] MAIN: Message size (" + strconv.Itoa(msgSize) + ") does not match specified 'size': " + strconv.Itoa(int(size))
			fmt.Println(custError)
			fmt.Println("!!!", receivedMessageRaw[:msgSize])
			con.Close()
			//return
		}
		receivedMsg := makeApiMessageOutOfBytes(receivedMessageRaw[:msgSize])
		//		fmt.Println("Parsed message into : ", receivedMsg.toString())

		switch receivedMsg.header.messageType {
		case dhtPUT:
			handlePut(receivedMsg.body.(*putBody))

		case dhtGET:
			if receivedMsg.header.size != 36 {
				custError := "[FAILURE] MAIN: Message size (" + strconv.Itoa(msgSize) + ") does not match expected size for a GET message"
				fmt.Println(custError)
				con.Close()
				return
			}
			answer := handleGet(receivedMsg.body.(*getBody))
			answerMessage := makeApiMessageOutOfAnswer(answer)

			_, err := con.Write(answerMessage.data)
			if err != nil {
				custError := "[FAILURE] MAIN:  Error while writing to connection: " + err.Error()
				fmt.Println(custError)
				panic(custError)
			}
			fmt.Println("[SUCCESS] MAIN: Written answer to connection")

		default:
			custError := "[FAILURE] MAIN: Message was of not specified type: " + strconv.Itoa(int(receivedMsg.header.messageType))
			fmt.Println(custError)
			con.Close()
			return
		}

		con.SetDeadline(time.Now().Add(time.Minute * 20)) //Timeout restarted
	}
}

func handleGet(body *getBody) DhtAnswer {
	//fmt.Println("handleGet has received :", body.toString())
	time.Sleep(1 * time.Second)
	v, ok := readMap(body.key)
	if ok {
		return DhtAnswer{
			success: true,
			key:     body.key,
			value:   v,
		}
	} else {
		return DhtAnswer{
			success: false,
			key:     body.key,
			value:   nil,
		}
	}
}

func handlePut(body *putBody) {
	//fmt.Println("handlePut has received :", body.toString())
	writeMap(body.key, body.value)
}

var testLock = sync.RWMutex{}

func readMap(key id) ([]byte, bool) {
	testLock.RLock()
	defer testLock.RUnlock()
	v, ok := testMap[key]
	return v, ok
}

func writeMap(key id, value []byte) {
	testLock.Lock()
	defer testLock.Unlock()
	testMap[key] = value
}

var testMap map[id][]byte
