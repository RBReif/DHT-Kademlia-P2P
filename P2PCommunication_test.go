package main

import (
	"crypto/rand"
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
	_, err := rand.Read(randByteArray)
	if err != nil {
		t.Errorf("Could not generate random id")
	}
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

// Test if pingNode successfully sends a PING request
func TestPingNode(t *testing.T) {

	if pingNode(thisNode.thisPeer, thisNode.thisPeer) != false {
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
	go pingNode(thisNode.thisPeer, thisNode.thisPeer)

	_, err = ln.Accept()
	if err != nil {
		t.Errorf("Error while listening for PING")
	}
	// else: no error --> PING successfully received

}
