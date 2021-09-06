package main

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestPingCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i
	ping1 := makeP2PMessageOutOfBody(nil, KDM_PING)
	fmt.Println("Ping1: ", ping1.toString())
	if ping1.body != nil {
		t.Errorf("Body of KDM_PING message is not nil")
	}

	fmt.Println(ping1.data)

	ping2 := makeP2PMessageOutOfBytes(ping1.data)
	fmt.Println("Ping2: ", ping2.toString())

	if ping1.header.size != ping2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if ping1.header.messageType != ping2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(ping1.header.nonce, ping2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(ping1.header.senderPeer.id, ping2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if ping1.header.senderPeer.port != ping2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if ping1.header.senderPeer.ip != ping2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}

}

func TestPongCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i
	pong1 := makeP2PMessageOutOfBody(nil, KDM_PONG)
	fmt.Println("Pong1: ", pong1.toString())
	if pong1.body != nil {
		t.Errorf("Body of KDM_PING message is not nil")
	}

	fmt.Println(pong1.data)

	pong2 := makeP2PMessageOutOfBytes(pong1.data)
	fmt.Println("Pong2: ", pong2.toString())

	if pong1.header.size != pong2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if pong1.header.messageType != pong2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(pong1.header.nonce, pong2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(pong1.header.senderPeer.id, pong2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if pong1.header.senderPeer.port != pong2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if pong1.header.senderPeer.ip != pong2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}

}

func TestStoreCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	idy := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idy); err != nil {
		panic(err.Error())
	}
	var i2 id
	copy(i2[:], idy)

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	storeBdy := kdmStoreBody{
		key:   i2,
		value: value,
	}

	kdmStore1 := makeP2PMessageOutOfBody(&storeBdy, KDM_STORE)
	fmt.Println("KDM_Store1: ", kdmStore1.toString())
	if kdmStore1.body == nil {
		t.Errorf("Body of KDM_Store message is  nil")
	}

	fmt.Println(kdmStore1.data)

	kdmStore2 := makeP2PMessageOutOfBytes(kdmStore1.data)
	fmt.Println("KDM_Store2: ", kdmStore2.toString())

	if kdmStore1.header.size != kdmStore2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if kdmStore1.header.messageType != kdmStore2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(kdmStore1.header.nonce, kdmStore2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(kdmStore1.header.senderPeer.id, kdmStore2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if kdmStore1.header.senderPeer.port != kdmStore2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if kdmStore1.header.senderPeer.ip != kdmStore2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}
	if !reflect.DeepEqual(kdmStore1.body, kdmStore2.body) {
		t.Errorf("Parsing of KDMStore body does not work")

	}
}

func TestFindNodeCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	idy := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idy); err != nil {
		panic(err.Error())
	}
	var i2 id
	copy(i2[:], idy)

	findNodeBdy := kdmFindNodeBody{
		id: i2,
	}

	findNode1 := makeP2PMessageOutOfBody(&findNodeBdy, KDM_FIND_NODE)
	fmt.Println("FindNode1: ", findNode1.toString())
	if findNode1.body == nil {
		t.Errorf("Body of FindNode message is  nil")
	}

	fmt.Println(findNode1.data)

	findNode2 := makeP2PMessageOutOfBytes(findNode1.data)
	fmt.Println("FindNode2: ", findNode2.toString())

	if findNode1.header.size != findNode2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if findNode1.header.messageType != findNode2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(findNode1.header.nonce, findNode2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(findNode1.header.senderPeer.id, findNode2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if findNode1.header.senderPeer.port != findNode2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if findNode1.header.senderPeer.ip != findNode2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}
	if !reflect.DeepEqual(findNode1.body, findNode2.body) {
		t.Errorf("Parsing of FindNode body does not work")

	}
}

func TestFindNodeAnswerCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	var ps []peer
	for i := 0; i < 5; i++ {
		idx := make([]byte, SIZE_OF_ID)
		if _, err := rand.Read(idx); err != nil {
			panic(err.Error())
		}
		var i2 id
		copy(i2[:], idx)

		p := peer{
			ip:   "1.1.1." + strconv.Itoa(i),
			port: uint16(3 * i),
			id:   i2,
		}

		ps = append(ps, p)

	}

	findNodeBdy := kdmFindNodeAnswerBody{
		answerPeers: ps,
	}

	findNodeAnswer := makeP2PMessageOutOfBody(&findNodeBdy, KDM_FIND_NODE_ANSWER)
	fmt.Println("FindNodeAnswer1: ", findNodeAnswer.toString())
	if findNodeAnswer.body == nil {
		t.Errorf("Body of FindNodeAnswer message is  nil")
	}

	fmt.Println(findNodeAnswer.data)

	findNodeAnswer2 := makeP2PMessageOutOfBytes(findNodeAnswer.data)
	fmt.Println("FindNodeAnswer2: ", findNodeAnswer2.toString())

	if findNodeAnswer.header.size != findNodeAnswer2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if findNodeAnswer.header.messageType != findNodeAnswer2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(findNodeAnswer.header.nonce, findNodeAnswer2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(findNodeAnswer.header.senderPeer.id, findNodeAnswer2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if findNodeAnswer.header.senderPeer.port != findNodeAnswer2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if findNodeAnswer.header.senderPeer.ip != findNodeAnswer2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}
	if !reflect.DeepEqual(findNodeAnswer.body, findNodeAnswer2.body) {
		t.Errorf("Parsing of FindNodeAnswer body does not work")

	}
}

func TestFindValueCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	idy := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idy); err != nil {
		panic(err.Error())
	}
	var i2 id
	copy(i2[:], idy)

	findValueBody := kdmFindValueBody{
		id: i2,
	}

	findValue1 := makeP2PMessageOutOfBody(&findValueBody, KDM_FIND_VALUE)
	fmt.Println("FindValue1: ", findValue1.toString())
	if findValue1.body == nil {
		t.Errorf("Body of FindValue message is  nil")
	}

	fmt.Println(findValue1.data)

	findValue2 := makeP2PMessageOutOfBytes(findValue1.data)
	fmt.Println("FindValue2: ", findValue2.toString())

	if findValue1.header.size != findValue2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if findValue1.header.messageType != findValue2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(findValue1.header.nonce, findValue2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(findValue1.header.senderPeer.id, findValue2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if findValue1.header.senderPeer.port != findValue2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if findValue1.header.senderPeer.ip != findValue2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}
	if !reflect.DeepEqual(findValue1.body, findValue2.body) {
		t.Errorf("Parsing of FindValue body does not work")

	}
}

func TestFoundValueCodingAndDecoding(t *testing.T) {
	thisNode.thisPeer.ip = "1.4.2.3"
	thisNode.thisPeer.port = 30
	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	idy := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idy); err != nil {
		panic(err.Error())
	}
	var key id
	copy(key[:], idy)

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}

	foundValueBdy := kdmFoundValueBody{
		key:   key,
		value: value,
	}

	foundValue1 := makeP2PMessageOutOfBody(&foundValueBdy, KDM_FOUND_VALUE)
	fmt.Println("FoundValue1: ", foundValue1.toString())
	if foundValue1.body == nil {
		t.Errorf("Body of FoundValue message is  nil")
	}

	fmt.Println(foundValue1.data)

	foundValue2 := makeP2PMessageOutOfBytes(foundValue1.data)
	fmt.Println("FindNode2: ", foundValue2.toString())

	if foundValue1.header.size != foundValue2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if foundValue1.header.messageType != foundValue2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(foundValue1.header.nonce, foundValue2.header.nonce) {
		t.Errorf("Parsing of Header nonce does not work")
	}
	if !reflect.DeepEqual(foundValue1.header.senderPeer.id, foundValue2.header.senderPeer.id) {
		t.Errorf("Parsing of Header sender Peer ID does not work")
	}
	if foundValue1.header.senderPeer.port != foundValue2.header.senderPeer.port {
		t.Errorf("Parsing of Header sender Peer port does not work")
	}
	if foundValue1.header.senderPeer.ip != foundValue2.header.senderPeer.ip {
		t.Errorf("Parsing of Header sender Peer IP does not work")
	}
	if !reflect.DeepEqual(foundValue1.body, foundValue2.body) {
		t.Errorf("Parsing of FoundValue body does not work")

	}
}

func TestParsePeerToByte(t *testing.T) {
	randIdBytes := make([]byte, SIZE_OF_ID)
	rand.Read(randIdBytes)
	var randId id
	for i := 0; i < SIZE_OF_ID; i++ {
		randId[i] = randIdBytes[i]
	}

	var peer = peer{"127.0.0.1", 1234, randId}

	var bytes = peerToByte(peer)

	if len(bytes) != 50 {
		t.Errorf("Peer has to have length 50")
	}

	// check if first SIZE_OF_ID bytes are equal to id
	for i := 0; i < SIZE_OF_ID; i++ {
		if bytes[i] != randIdBytes[i] {
			t.Errorf("First " + fmt.Sprint(SIZE_OF_ID) + " Bytes have to be equal to Id")
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
