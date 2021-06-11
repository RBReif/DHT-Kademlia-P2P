package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
)

// TODO: move k to singleton or something
var k int
var a int
var n peer

const KDM_PING uint16 = 654
const KDM_PONG uint16 = 655
const KDM_STORE uint16 = 656
const KDM_FIND_NODE uint16 = 657
const KDM_FIND_VALUE uint16 = 658

type id [20]byte

type peer struct {
	ip   string
	port int
	id   id

	kBuckets [160][]peer
}

func startMessageDispatcher() {

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// TODO: handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// TODO: handle error
		}

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {

	var message message
	readHeader(conn, &message)
	n.onMessageReceived(message)

	// switch according to message type
	switch message.messageType {
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
		// TODO
		return
	case KDM_FIND_VALUE:
		// TODO
		return
	}

}

func (node *peer) onMessageReceived(message message) {
	//first we need to find out which is the responsible Bucket
	indexResponsibleBucket := node.findIndexOfResponsibleBucket(message.sender.id)

	// if id of sender exists in kBuckets, move to tail of kBuckets
	if index, inBucket := isIdInKBucket(node.kBuckets[0], message.sender.id); inBucket {
		moveToTail(node.kBuckets[0], index)
	} else {
		// if kBuckets has fewer than k entries, insert id to kBuckets
		if len(node.kBuckets[indexResponsibleBucket]) < maxSizeOfBucket(indexResponsibleBucket) {
			node.kBuckets[indexResponsibleBucket] = append(node.kBuckets[indexResponsibleBucket], message.sender)
		} else {
			// else ping least-recently seen node
			nodeActive := pingNode(message.sender)

			// if node not responding, remove node and insert the new one
			if !nodeActive {
				node.kBuckets[indexResponsibleBucket] = append(node.kBuckets[indexResponsibleBucket][:index], node.kBuckets[indexResponsibleBucket][index+1:]...)
				node.kBuckets[indexResponsibleBucket] = append(node.kBuckets[indexResponsibleBucket], message.sender)
			} else {
				// else move node to tail and discard the new one
				moveToTail(node.kBuckets[indexResponsibleBucket], index)
			}
		}

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
	readHeader(c, &message)
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

func (p *peer) nodeLookup(key id) {
	var closestPeersOld []peer
	for {
		closestPeersNew := p.findNumberOfClosestPeersOnNode(key, a)
		if !wasNewPeerAdded(closestPeersOld, closestPeersNew) {
			break
		}

		for _, p := range closestPeersNew {
			p.FIND_NODE(key)
		}
		closestPeersOld = closestPeersNew
	}
}

func (p *peer) FIND_NODE(key id) {
	closestPeers := p.findNumberOfClosestPeersOnNode(key, a)
	answerMessage := makeFIND_NODEanswer(closestPeers)
	fmt.Println(answerMessage) //todo replace with sending the answer
}
