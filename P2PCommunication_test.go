package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"testing"
)

func TestDistance(t *testing.T) {
	id1 := id{}
	id2 := id{}

	if distance(id1, id2) != id1 {
		t.Errorf("Distance of two empty ids has to be empty id")
	}
}

func TestToByte(t *testing.T) {
	// prepare random id
	randByteArray := make([]byte, SIZE_OF_ID)
	rand.Read(randByteArray)
	var randId id
	for i := 0; i < SIZE_OF_ID; i++ {
		randId[i] = randByteArray[i]
	}

	// test toByte function
	convertedArray := randId.toByte()
	for i := 0; i < SIZE_OF_ID; i++ {
		if convertedArray[i] != randByteArray[i] {
			t.Errorf("id.toByte() does not return correct byte array")
			break
		}
	}
}

func TestInitMakeHashtable(t *testing.T) {
	if n.hashTable != nil {
		t.Errorf("Hashtable must not be initialized before init function called")
	}

	n.init()

	if n.hashTable == nil {
		t.Errorf("Hashtable must not be nil after init function called")
	}

}

func TestPingNode(t *testing.T) {

	if pingNode(n.peer) != false {
		t.Errorf("Ping of unavailable Node has to be false")
	}

	// Listen for Ping request
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Errorf("Error while listening for tcp connection")
	}

	n.peer.ip = "127.0.0.1"
	n.peer.port = 8080

	// send Ping request
	go pingNode(n.peer)

	_, err = ln.Accept()
	if err != nil {
		t.Errorf("Error while listening for PING")
	} else {
		// no error --> PING successfully received
	}

}

func TestKDM_PING(t *testing.T) {

	n.peer.ip = "127.0.0.1"
	n.peer.port = 8080

	go n.startMessageDispatcher()

	c, err := net.Dial("tcp", n.peer.ip+":"+fmt.Sprint(n.peer.port))
	if err != nil {
		t.Errorf("Error opening TCP Connection")
	}
	pingMessage := makeMessageOutOfBody(nil, KDM_PING)
	pingMessage.header.senderPeer.port = 8081 // change port to port of test case
	sendMessage(pingMessage, n.peer)

	// receive KDM_PONG
	answerRaw := readMessage(c)
	answer := makeMessageOutOfBytes(answerRaw)
	if answer.header.messageType == KDM_PONG {
		// success
	} else {
		// failure
		t.Errorf("Received no KDM_PONG after sending KDM_PING")
	}

}
