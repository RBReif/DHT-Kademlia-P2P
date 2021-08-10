package p2p

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestPingPong(t *testing.T) {
	n.peer.ip = "1.4.2.3"
	n.peer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	n.peer.id = i
	ping1 := makeMessageOutOfBody(nil, KDM_PING)
	fmt.Println("Ping1: ", ping1.toString())
	if ping1.body != nil {
		t.Errorf("Body of KDM_PING message is not nil")
	}

	fmt.Println(ping1.data)

	ping2 := makeMessageOutOfBytes(ping1.data)
	fmt.Println("Ping2: ", ping2.toString())

	//t.Errorf("Distance of two empty ids has to be empty id")

}
