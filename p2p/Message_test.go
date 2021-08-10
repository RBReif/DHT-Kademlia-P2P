package p2p

import (
	"crypto/rand"
	"testing"
)

func TestParsePeerToByte(t *testing.T) {
	randIdBytes := make([]byte, SIZE_OF_ID)
	rand.Read(randIdBytes)
	var randId id
	for i := 0; i < SIZE_OF_ID; i++ {
		randId[i] = randIdBytes[i]
	}

	var peer = peer{"127.0.0.1", true, 1234, randId}

	var bytes = peerToByte(peer)

	if len(bytes) != 38 {
		t.Errorf("Peer has to have length 38")
	}

	// check if first 20 bytes are equal to id
	for i := 0; i < 20; i++ {
		if bytes[i] != randIdBytes[i] {
			t.Errorf("First 20 Bytes have to be equal to Id")
		}
	}

	// check if next 16 bytes are equal to ip
	// TODO

	// check if last 2 bytes are equal to port
	// TODO

	// check if parsing peer to bytes back to peer results in initial peer
	if peer != parseByteToPeer(peerToByte(peer)) {
		t.Errorf("Parsed peer to bytes and backwards has to be equal to initial peer")
	}

}
