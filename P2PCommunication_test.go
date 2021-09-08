package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
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
	if thisNode.hashTable.values != nil || thisNode.hashTable.expirations != nil {
		t.Errorf("Hashtable must not be initialized before init function called")
	}

	thisNode.init()

	if thisNode.hashTable.values == nil || thisNode.hashTable.expirations == nil {
		t.Errorf("Hashtable must not be nil after init function called")
	}

}

// Test if pingNode successfully sends a PING request
func TestPingNode(t *testing.T) {

	if pingNode(thisNode.thisPeer) != false {
		t.Errorf("Ping of unavailable Node has to be false")
	}

	// Listen for Ping request
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Errorf("Error while listening for tcp connection")
	}

	thisNode.thisPeer.ip = "127.0.0.1"
	thisNode.thisPeer.port = 8080

	// send Ping request
	go pingNode(thisNode.thisPeer)

	_, err = ln.Accept()
	if err != nil {
		t.Errorf("Error while listening for PING")
	} else {
		// no error --> PING successfully received
	}

}

// Test if node reacts correctly to PING message
func TestKDM_PING(t *testing.T) {

	thisNode.thisPeer.ip = "127.0.0.1"
	thisNode.thisPeer.port = 3011

	Conf.p2pIP = "127.0.0.1"
	Conf.p2pPort = 3011

	var wg sync.WaitGroup
	wg.Add(1)

	kBucket := kBucket{}
	kBucket = make([]peer, 0)

	thisNode.routingTree = routingTree{
		left:    nil,
		right:   nil,
		parent:  nil,
		prefix:  "",
		kBucket: kBucket,
	}

	go startP2PMessageDispatcher(&wg)
	c, err := net.Dial("tcp", thisNode.thisPeer.ip+":"+fmt.Sprint(thisNode.thisPeer.port))
	if err != nil {
		t.Errorf("Error opening TCP Connection: " + err.Error())
	}
	pingMessage := makeP2PMessageOutOfBody(nil, KDM_PING)
	tmp, _ := strconv.Atoi(strings.Split(c.LocalAddr().String(), ":")[1])
	pingMessage.header.senderPeer.port = uint16(tmp) // change port to port of test case
	pingMessage.data = pingMessage.header.decodeHeaderToBytes()
	sendP2PMessage(pingMessage, thisNode.thisPeer)
	fmt.Println("Sent Ping Message")
	l, err := net.Listen("tcp", Conf.p2pIP+":"+strconv.Itoa(tmp))
	conn, err := l.Accept()

	// receive KDM_PONG
	answer := readMessage(conn)
	//answer := makeP2PMessageOutOfBytes(answerRaw)
	fmt.Println("Received answer: ", answer.toString())
	if answer.header.messageType == KDM_PONG {
		// success
	} else {
		// failure
		t.Errorf("Received no KDM_PONG after sending KDM_PING")
	}
	wg.Done()
	fmt.Println("finished")

}
