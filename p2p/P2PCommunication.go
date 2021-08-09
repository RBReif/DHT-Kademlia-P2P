package p2p

import (
	//"encoding/binary"
	"fmt"
	"net"
)

// TODO: move k to singleton or something
var k int
var a int
var n localNode

// TODO: which maximum size should data have?
var hashTable map[id][]byte

type id [SIZE_OF_ID]byte

func (id id) toByte() []byte {
	var result []byte
	for i := 0; i < SIZE_OF_ID; i++ {
		result = append(result, id[i])
	}
	return result
}

type peer struct {
	ip     string
	isIpv4 bool
	port   uint16
	id     id
}

type localNode struct {
	peer     peer
	kBuckets [160][]peer
}

func (thisNode *localNode) init() {
	thisNode.startMessageDispatcher()
	hashTable = make(map[id][]byte)
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

	m := readMessage(conn) //todo read whole message
	n.updateKBucketPeer(m.header.senderPeer)

	// switch according to m type
	switch m.header.messageType {
	case KDM_PING: // ping
		// respond with KDM_PONG
		pongMessage := makeMessage(nil, KDM_PONG)
		sendMessage(pongMessage, m.header.senderPeer)
		err := conn.Close()
		if err != nil {
			return
		}
	case KDM_STORE:
		// write (key, value) to hashTable
		hashTable[m.body.(*kdmStoreBody).key] = m.body.(*kdmStoreBody).value
		return
	case KDM_FIND_NODE:
		var key id
		copy(key[:], m.data[44:64])
		answerBody := thisNode.FIND_NODE(key)
		answer := makeMessage(&answerBody, KDM_FIND_NODE_ANSWER)

		sendMessage(answer, m.header.senderPeer)
		return

	case KDM_FIND_NODE_ANSWER:
		newPeers := m.body.(*kdmFindNodeAnswerBody).answerPeers
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
	pingMessage := makeMessage(nil, KDM_PING)
	sendMessage(pingMessage, node)
	// receive KDM_PONG
	answer := readMessage(c)
	if answer.header.messageType == KDM_PONG {
		return true
	}

	return false

}

func (thisNode *localNode) nodeLookup(key id) {
	var closestPeersOld []peer
	for {
		closestPeersNew := thisNode.findNumberOfClosestPeersOnNode(key, a)
		if !wasAnyNewPeerAdded(closestPeersOld, closestPeersNew) {
			break
		}
		//todo maybe change procedure to also call rpc on newly added nodes, that are farer away then the ones queried in round before
		//todo maybe collect the answers first and use them during nodeLookup before updating the kBuckets
		for _, p := range closestPeersNew {
			if wasANewPeerAdded(closestPeersOld, p) {
				msgBody := kdmFindNodeBody{
					id: key,
				}
				m := makeMessage(&msgBody, KDM_FIND_NODE)
				sendMessage(m, p)
			}
		}
		closestPeersOld = closestPeersNew
		//todo wait , for how long?
	}
}

func (thisNode *localNode) FIND_NODE(key id) kdmFindNodeAnswerBody {
	closestPeers := thisNode.findNumberOfClosestPeersOnNode(key, k)
	answerBody := kdmFindNodeAnswerBody{answerPeers: closestPeers}
	fmt.Println(answerBody)
	return answerBody
}
