package main

import (
	//"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var thisNode localNode

type localNode struct {
	thisPeer    peer
	routingTree routingTree
	// TODO: which maximum size should data have?
	hashTable hashTable
}

type hashTable struct {
	values            map[id][]byte
	expirations       map[id]time.Time
	republishingTimes map[id]time.Time
	sync.RWMutex
}

func (hashTable *hashTable) read(key id) ([]byte, bool) {
	hashTable.RLock()
	defer hashTable.RUnlock()
	var value, existing = hashTable.values[key]
	return value, existing
}

func (hashTable *hashTable) write(key id, value []byte, expiration time.Time, republishingTime time.Time) {
	hashTable.Lock()
	defer hashTable.Unlock()
	hashTable.values[key] = value
	hashTable.expirations[key] = expiration
	hashTable.republishingTimes[key] = republishingTime
}

func (hashTable *hashTable) republishKeys() {
	hashTable.Lock()
	defer hashTable.Unlock()
	for key, value := range hashTable.republishingTimes {
		if time.Now().After(value) {
			// TODO republish keys
			print("DEBUG: Republishing" + fmt.Sprint(key)) // TODO: delete this line
		}
	}
}

// Deletes all key/value-pairs which are expired
func (hashTable *hashTable) expireKeys() {
	hashTable.Lock()
	defer hashTable.Unlock()
	for key, value := range hashTable.expirations {
		if time.Now().After(value) {
			delete(hashTable.values, key)
			delete(hashTable.expirations, key)
		}
	}
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

func (id *id) startsWith(prefix string) bool {
	idString := ""
	for _, byte := range id {
		idString += fmt.Sprintf("%08b", byte)
	}
	return strings.HasPrefix(idString, prefix)
}

func (id id) toByte() []byte {
	var result []byte
	for i := 0; i < SIZE_OF_ID; i++ {
		result = append(result, id[i])
	}
	return result
}

func (thisNode *localNode) init() {
	thisNode.hashTable.values = make(map[id][]byte)
}

func startP2PMessageDispatcher(wg *sync.WaitGroup) {
	defer wg.Done()

	l, err := net.Listen("tcp", Conf.p2pIP+":"+strconv.Itoa(int(Conf.p2pPort)))
	if err != nil {
		custError := "[FAILURE] MAIN: Error while listening for connection at" + Conf.p2pIP + ": " + strconv.Itoa(int(Conf.p2pPort)) + " - " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	defer l.Close()
	fmt.Println("[SUCCESS] MAIN: P2PMessageDispatcher Listening on ", Conf.p2pIP, ": ", Conf.p2pPort)
	for {
		conn, err := l.Accept()
		if err != nil {
			custError := "[FAILURE] MAIN: Error while accepting: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		fmt.Println("[SUCCESS] MAIN: New Connection established, ", conn)
		conn.SetDeadline(time.Now().Add(time.Minute * 20)) //Set Timeout

		go handleP2PConnection(conn)

	}

}

func handleP2PConnection(conn net.Conn) {

	mRaw := readMessage(conn) //todo readMap whole message
	m := makeP2PMessageOutOfBytes(mRaw)
	thisNode.updateKBucket(m.header.senderPeer)

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
		thisNode.hashTable.write(m.body.(*kdmStoreBody).key, m.body.(*kdmStoreBody).value, time.Now().Add(time.Duration(Conf.maxTTL)*time.Second), time.Now().Add(time.Duration(Conf.republishingTime)*time.Second))
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
			thisNode.updateKBucket(newPeers[i])
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

// checks every second if keys are expired or should be republished
func startTimers() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			// expire keys
			thisNode.hashTable.expireKeys()

			// republish keys
			thisNode.hashTable.republishKeys()
		}
	}
}
