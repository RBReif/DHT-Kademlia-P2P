package main

import (
	//"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
)

var thisNode localNode

type localNode struct {
	thisPeer peer
	kBuckets [160][]peer
	// TODO: which maximum size should data have?
	hashTable hashTable
}

type hashTable struct {
	values map[id][]byte
	sync.RWMutex
}

func (hashTable *hashTable) read(key id) ([]byte, bool) {
	hashTable.RLock()
	defer hashTable.RUnlock()
	var value, existing = hashTable.values[key]
	return value, existing
}

func (hashTable *hashTable) write(key id, value []byte) {
	hashTable.Lock()
	defer hashTable.Unlock()
	hashTable.values[key] = value
}

type peer struct {
	ip     string
	isIpv4 bool
	port   uint16
	id     id
}

func (p *peer) toString() string {
	return "" + p.ip + ":" + strconv.Itoa(int(p.port)) + " (" + bytesToString(p.id.toByte()) + ")"
}

type id [SIZE_OF_ID]byte

func (id id) toByte() []byte {
	var result []byte
	for i := 0; i < SIZE_OF_ID; i++ {
		result = append(result, id[i])
	}
	return result
}

func (thisNode *localNode) init() {

	// TODO: is starting of MessageDispatcher sensible here? thisNode.startMessageDispatcher()
	thisNode.hashTable.values = make(map[id][]byte)
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

	mRaw := readMessage(conn) //todo readMap whole message
	m := makeP2PMessageOutOfBytes(mRaw)
	thisNode.updateKBucketPeer(m.header.senderPeer)

	// switch according to m type
	switch m.header.messageType {
	case KDM_PING: // ping
		// respond with KDM_PONG
		pongMessage := makeP2PMessageOutOfBody(nil, KDM_PONG)
		sendP2PMessage(pongMessage, m.header.senderPeer)
		err := conn.Close()
		if err != nil {
			return
		}
	case KDM_STORE:
		// write (key, value) to hashTable
		thisNode.hashTable.write(m.body.(*kdmStoreBody).key, m.body.(*kdmStoreBody).value)
		return
	case KDM_FIND_NODE:
		var key id
		copy(key[:], m.data[44:64])
		answerBody := thisNode.FIND_NODE(key)
		answer := makeP2PMessageOutOfBody(&answerBody, KDM_FIND_NODE_ANSWER)

		sendP2PMessage(answer, m.header.senderPeer)
		return

	case KDM_FIND_NODE_ANSWER:
		newPeers := m.body.(*kdmFindNodeAnswerBody).answerPeers
		for i := 0; i < len(newPeers); i++ {
			thisNode.updateKBucketPeer(newPeers[i])
		}
		return

	case KDM_FIND_VALUE:
		var key id
		copy(key[:], m.data[44:64])
		var value, existing = thisNode.hashTable.read(key)
		if existing {
			// reply with value
			answerBody := kdmFoundValueBody{value: value}
			answer := makeP2PMessageOutOfBody(&answerBody, KDM_FOUND_VALUE)
			sendP2PMessage(answer, m.header.senderPeer)
		} else {
			// same behavior as KDM_FIND_NODE
			answerBody := thisNode.FIND_NODE(key)
			answer := makeP2PMessageOutOfBody(&answerBody, KDM_FIND_NODE_ANSWER)

			sendP2PMessage(answer, m.header.senderPeer)
		}
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

	c, err := net.Dial("tcp", node.ip+":"+fmt.Sprint(node.port))
	if err != nil {
		fmt.Println(err)
		return false
	}
	pingMessage := makeP2PMessageOutOfBody(nil, KDM_PING)
	sendP2PMessage(pingMessage, node)
	// receive KDM_PONG
	answerRaw := readMessage(c)
	answer := makeP2PMessageOutOfBytes(answerRaw)
	if answer.header.messageType == KDM_PONG {
		return true
	}

	return false

}

func (thisNode *localNode) nodeLookup(key id) {
	var closestPeersOld []peer
	for {
		closestPeersNew := thisNode.findNumberOfClosestPeersOnNode(key, Conf.a)
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
				m := makeP2PMessageOutOfBody(&msgBody, KDM_FIND_NODE)
				sendP2PMessage(m, p)
			}
		}
		closestPeersOld = closestPeersNew
		//todo wait , for how long?
	}
}

func (thisNode *localNode) FIND_NODE(key id) kdmFindNodeAnswerBody {
	closestPeers := thisNode.findNumberOfClosestPeersOnNode(key, Conf.k)
	answerBody := kdmFindNodeAnswerBody{answerPeers: closestPeers}
	fmt.Println(answerBody)
	return answerBody
}
