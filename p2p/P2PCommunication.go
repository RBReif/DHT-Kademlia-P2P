package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
)

// TODO: move k to singleton or something
var k int
var a int
var n localNode

const KDM_PING uint16 = 654
const KDM_PONG uint16 = 655
const KDM_STORE uint16 = 656
const KDM_FIND_NODE uint16 = 657
const KDM_FIND_NODE_ANSWER uint16 = 658
const KDM_FIND_VALUE uint16 = 659
const KDM_FIND_VALUE_ANSWER uint16 = 660

const SIZE_OF_IP int = 16
const SIZE_OF_PORT int = 4
const SIZE_OF_ID int = 20
const SIZE_OF_KEY int = 20

type id [20]byte

type peer struct {
	ip   string
	port uint32
	id   id
}

type localNode struct {
	peer     peer
	kBuckets [160][]peer
}

func (thisNode *localNode) startMessageDispatcher() {

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// TODO: handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// TODO: handle error
		}

		go thisNode.handleConnection(conn)

	}

}

func (thisNode *localNode) handleConnection(conn net.Conn) {

	var m message
	readMessage(conn, &m) //todo read whole message
	n.updateKBucketPeer(m.sender)

	// switch according to m type
	switch m.messageType {
	case KDM_PING: // ping
		// respond with KDM_PONG
		pongNode(conn)
		err := conn.Close()
		if err != nil {
			return
		}
	case KDM_STORE:
		// TODO
		return
	case KDM_FIND_NODE:
		var key id
		copy(key[:], m.data[44:64])
		answerData := thisNode.FIND_NODE(key)
		answer := message{
			data:        answerData,
			sender:      thisNode.peer,
			receiver:    m.sender,
			size:        uint16(len(answerData)),
			messageType: KDM_FIND_NODE_ANSWER,
		}
		sendMessage(answer)
		return

	case KDM_FIND_NODE_ANSWER:
		newPeers := parseFIND_NODE_ANSWER(m)
		for i := 0; i < len(newPeers); i++ {
			thisNode.updateKBucketPeer(newPeers[i])
		}
		return

	case KDM_FIND_VALUE:
		// TODO
		return
	}

}

// distance function of kademlia
func distance(id1 id, id2 id) id {

	xor := id{}

	for i := 0; i < len(id1); i++ {
		xor[i] = id1[i] ^ id2[i]

	}
	return xor
}

// probes a node to check if it is online
func pingNode(node peer) bool {

	c, err := net.Dial("tcp", node.ip+"node.port")
	if err != nil {
		fmt.Println(err)
		return false
	}

	var pingRequest [4]byte

	// write size
	binary.BigEndian.PutUint16(pingRequest[0:], uint16(4))

	// write KDM_PING
	binary.BigEndian.PutUint16(pingRequest[2:], KDM_PING)

	// send KDM_PING
	_, err = c.Write(pingRequest[:])
	if err != nil {
		fmt.Println(err)
		return false
	}

	// receive KDM_PONG
	var message message
	readMessage(c, &message)
	if message.messageType == KDM_PONG {
		return true
	}

	return false

}

func pongNode(c net.Conn) {

	var pongResponse [4]byte

	// write size
	binary.BigEndian.PutUint16(pongResponse[0:], uint16(4))

	// write KDM_PING
	binary.BigEndian.PutUint16(pongResponse[2:], KDM_PONG)

	// send KDM_PING
	_, err := c.Write(pongResponse[:])
	if err != nil {
		fmt.Println(err)
		return
	}

}

func (thisNode *localNode) nodeLookup(key id) {
	var closestPeersOld []peer
	for {
		closestPeersNew := thisNode.findNumberOfClosestPeersOnNode(key, a)
		if !wasAnyNewPeerAdded(closestPeersOld, closestPeersNew) {
			break
		}
		//todo maybe change procedure to also call rpc on newly added nodes, that are farer away then the ones queried in round before
		for _, p := range closestPeersNew {
			if wasANewPeerAdded(closestPeersOld, p) {
				m := message{
					sender:      thisNode.peer,
					receiver:    p,
					size:        uint16(SIZE_OF_ID + SIZE_OF_IP + SIZE_OF_PORT + 4 + SIZE_OF_KEY),
					messageType: KDM_FIND_NODE,
					data:        makeFIND_NODEmessage(key),
				}
				sendMessage(m)
			}
		}
		closestPeersOld = closestPeersNew
	}
}

func (thisNode *localNode) FIND_NODE(key id) []byte {
	closestPeers := thisNode.findNumberOfClosestPeersOnNode(key, a)
	answerMessage := makeFIND_NODE_ANSWERmessage(closestPeers)
	fmt.Println(answerMessage)
	return answerMessage
}
