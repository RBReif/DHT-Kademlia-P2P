package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// local peer in the network
var thisNode localNode

// struct which represents the local peer in the network
type localNode struct {
	thisPeer    peer
	routingTree routingTree
	hashTable   hashTable
}

// struct which represents the data storage
// also contains information for expiration and republishing times
type hashTable struct {
	values            map[id][]byte
	expirations       map[id]time.Time
	republishingTimes map[id]time.Time
	sync.RWMutex
}

// reads value for given key from the local data storage
// returns the value or nil and a boolean if the value was found
func (hashTable *hashTable) read(key id) ([]byte, bool) {
	hashTable.RLock()
	defer hashTable.RUnlock()
	var value, existing = hashTable.values[key]
	return value, existing
}

// writes <key, value>-pair to the local data storage
func (hashTable *hashTable) write(key id, value []byte, expiration time.Time, republishingTime time.Time) {
	hashTable.Lock()
	defer hashTable.Unlock()
	hashTable.values[key] = value
	hashTable.expirations[key] = expiration
	hashTable.republishingTimes[key] = republishingTime
	log.Debug("WE HAVE WRITTEN KEY_VALUE PAIR TO ", Conf.p2pPort, " :", key[:10], "  - ", value, " (ttl ", expiration, ")")
}

// checks for all stored <key, value>-pairs if they need to be republished to the network and republishes them if so
func (hashTable *hashTable) republishKeys() {
	hashTable.Lock()
	defer hashTable.Unlock()
	for key, value := range hashTable.republishingTimes {
		if time.Now().After(value) { // if republishingTime lies in the past
			log.Debug("Republishing: " + fmt.Sprint(key))
			store(key, hashTable.values[key], uint16(time.Until(hashTable.expirations[key]))) // republish
		}
	}
}

// removes all key/value-pairs which are expired
func (hashTable *hashTable) expireKeys() {
	hashTable.Lock()
	defer hashTable.Unlock()
	for key, value := range hashTable.expirations {
		if time.Now().After(value) { // if expiration time lies in the past
			// remove <key, value>-pair completely from the hashTable
			delete(hashTable.values, key)
			delete(hashTable.expirations, key)
			delete(hashTable.republishingTimes, key)
		}
	}
}

// struct which represents a peer in the network as triple of <ip, port, id>
type peer struct {
	ip   string
	port uint16
	id   id
}

func (p *peer) toString() string {
	return "" + p.ip + ":" + strconv.Itoa(int(p.port)) + " (" + bytesToString(p.id.toByte()[:10]) + ")"
}

// type which represents the id of a peer
type id [SIZE_OF_ID]byte

// helper function for checking if id starts with given prefix
func (id *id) startsWith(prefix string) bool {
	idString := ""
	for _, byte := range id {
		idString += fmt.Sprintf("%08b", byte)
	}
	return strings.HasPrefix(idString, prefix)
}

// converts id to byte slice
func (id id) toByte() []byte {
	var result []byte
	for i := 0; i < SIZE_OF_ID; i++ {
		result = append(result, id[i])
	}
	return result
}

// initializes P2P-Communication
func initializeP2PCommunication() {
	//first we read the private key
	priv, err := ioutil.ReadFile(Conf.HostKeyFile)
	if err != nil {
		log.Error("Error while reading File: ", err)
	}
	log.Debug("   Bytes from private key pem file: ", priv[:10], "...", priv[140:150], "...")
	block, _ := pem.Decode([]byte(priv))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("[FAILURE] failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	//now we generate the corresponding public key
	publicKeyDer, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}
	pubKeyPem := string(pem.EncodeToMemory(&pubKeyBlock))
	log.Debug("   public key generated: ", pubKeyPem[:50], "...")

	log.Debug("   Public Key as bytes: ", publicKeyDer[:10], "...", publicKeyDer[140:150], "...")

	//now we calculate the sha256 hash sum to retreive our ID
	h := sha256.New()
	h.Write(publicKeyDer)
	newIDbytes := h.Sum(nil)
	var newID id
	copy(newID[:], newIDbytes)
	log.Info("[SUCCESSS] Our generated ID: ", newIDbytes)
	thisNode.thisPeer.ip = Conf.p2pIP
	thisNode.thisPeer.port = Conf.p2pPort
	thisNode.thisPeer.id = newID
	log.Info("[SUCCESS] Configured this peer: ", thisNode.thisPeer.toString())
	initialPeers := make([]peer, 3)
	initialPeers[0] = extractPeerAddressFromString(Conf.preConfPeer1)
	initialPeers[1] = extractPeerAddressFromString(Conf.preConfPeer2)
	initialPeers[2] = extractPeerAddressFromString(Conf.preConfPeer3)

	kBucket := make([]peer, 0)

	thisNode.routingTree = routingTree{
		left:    nil,
		right:   nil,
		parent:  nil,
		prefix:  "",
		kBucket: kBucket,
	}
	time.Sleep(1 * time.Second)
	for _, p := range initialPeers {
		if pingNode(p, thisNode.thisPeer) {
			thisNode.updateRoutingTable(p)
		}
	}

	thisNode.hashTable = hashTable{
		values:            make(map[id][]byte),
		expirations:       make(map[id]time.Time),
		republishingTimes: make(map[id]time.Time),

		RWMutex: sync.RWMutex{},
	}
	log.Info("[SUCCESS] FINISHED INITIALIZING OF P2P COMMUNICATION\n")
	time.Sleep(1 * time.Second)
	log.Debug(thisNode.thisPeer.port, "stores: ", thisNode.routingTree.toString())
}

/*
This function can be used to create a peer from the ip:port format and considers both IPv4 and IPv6 formats as specified
in the specification. The peer that is returned does not yet include an ID
*/
func extractPeerAddressFromString(line string) peer {
	result := peer{}
	var ip string
	var port int
	var err error
	if strings.Contains(line, "[") {
		ip = strings.Split(line, "]:")[0][1:]
		port, err = strconv.Atoi(strings.Split(line, "]:")[1])
	} else {
		ip = strings.Split(line, ":")[0]
		port, err = strconv.Atoi(strings.Split(line, ":")[1])
	}
	if err != nil {
		log.Fatal("Wrong configuration: the port of ", line, " is not an Integer")
	}
	result.ip = ip
	result.port = uint16(port)
	return result
}

// starts the message dispatcher for P2P-Communication
// listens for connections and delegates them to handleP2PConnection
func startP2PMessageDispatcher(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	// start listening on ip and port
	l, err := net.Listen("tcp", Conf.p2pIP+":"+strconv.Itoa(int(Conf.p2pPort)))
	if err != nil {
		log.Panic("[FAILURE] MAIN: Error while listening for connection at" + Conf.p2pIP + ": " + strconv.Itoa(int(Conf.p2pPort)) + " - " + err.Error())
	}
	defer l.Close()
	log.Info("[SUCCESS] MAIN: P2PMessageDispatcher Listening on ", Conf.p2pIP, ": ", Conf.p2pPort)

	go func() {
		for {
			// accept connections
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					// program canceled, no error
					return
				default:
					custError := "[FAILURE] MAIN: Error while accepting: " + err.Error()
					log.Panic(custError)
				}
			}
			log.Debug("[SUCCESS] MAIN: New Connection established, ", conn.LocalAddr(), " r:", conn.RemoteAddr())

			// set timeout
			err = conn.SetDeadline(time.Now().Add(time.Minute * 20))
			if err != nil {
				custError := "[FAILURE] MAIN: Error while setting Deadline: " + err.Error()
				log.Panic(custError)
			}

			// delegate handling of incoming connection
			go handleP2PConnection(conn)

		}
	}()

	// listen for program to stop
	for {
		select {
		case <-ctx.Done():
			log.Debug("[DEBUG] P2PMessageDispatcher received stoppingSignal")
			l.Close()
			return
		}
	}

}

// handles incoming connection based on message type
func handleP2PConnection(conn net.Conn) {

	m := readMessage(conn)
	if m != nil {
		bdyStrg := ""
		if m.body != nil {
			bdyStrg = m.body.toString()
		}
		log.Info(thisNode.thisPeer.ip, ":", thisNode.thisPeer.port, " has received this message: ", m.header.toString(), " : ", bdyStrg)

		// update routing table
		thisNode.updateRoutingTable(m.header.senderPeer)

		// switch according to message type
		switch m.header.messageType {
		case KDM_PING:
			// respond with KDM_PONG
			pongMessage := makeP2PMessageOutOfBody(nil, KDM_PONG)
			sendP2PMessage(pongMessage, m.header.senderPeer)
			_, err := conn.Write(pongMessage.data)
			if err != nil {
				return
			}
			err = conn.Close()
			if err != nil {
				return
			}
		case KDM_STORE:
			// write <key, value>-pair to hashTable
			ttl := int(m.body.(*kdmStoreBody).ttl)
			if ttl > Conf.maxTTL {
				ttl = Conf.maxTTL
			}
			thisNode.hashTable.write(m.body.(*kdmStoreBody).key, m.body.(*kdmStoreBody).value, time.Now().Add(time.Duration(ttl)*time.Second), time.Now().Add(time.Duration(REPUBLISH_TIME)*time.Second))
			return

		case KDM_FOUND_VALUE:
			// write found <key, value>-pair to hashTable
			thisNode.hashTable.write(m.body.(*kdmFoundValueBody).key, m.body.(*kdmFoundValueBody).value, time.Now().Add(time.Duration(15)*time.Second), time.Now().Add(time.Duration(REPUBLISH_TIME)*time.Second))

		case KDM_FIND_NODE:
			key := m.body.(*kdmFindNodeBody).id

			// find k closest nodes to given id on local node and return them to sender
			answerBody := thisNode.FIND_NODE(key)
			answer := makeP2PMessageOutOfBody(&answerBody, KDM_FIND_NODE_ANSWER)
			sendP2PMessage(answer, m.header.senderPeer)
			return

		case KDM_FIND_NODE_ANSWER:
			// extract found peers and update routingTable accordingly
			newPeers := m.body.(*kdmFindNodeAnswerBody).answerPeers
			for i := 0; i < len(newPeers); i++ {
				thisNode.updateRoutingTable(newPeers[i])
			}
			return

		case KDM_FIND_VALUE:
			key := m.body.(*kdmFindValueBody).id

			// look for value to given key in local hashTable
			var value, existing = thisNode.hashTable.read(key)
			if existing {
				// reply with value
				answerBody := kdmFoundValueBody{value: value, key: key}
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
func pingNode(receiverPeer peer, senderPeer peer) bool {

	c, err := net.Dial("tcp", receiverPeer.ip+":"+fmt.Sprint(receiverPeer.port))
	if err != nil {
		log.Error(err)
		return false
	}

	pingMessage := p2pMessage{}
	pingMessage.header.messageType = KDM_PING

	pingMessage.header.senderPeer = senderPeer
	nonce := make([]byte, SIZE_OF_NONCE)
	if _, err := rand.Read(nonce); err != nil {
		panic(err.Error())
	}
	pingMessage.header.nonce = nonce
	log.Debug(pingMessage.header.messageType, pingMessage.header.nonce)
	log.Debug(pingMessage.header.senderPeer)
	pingMessage.header.size = uint16(SIZE_OF_HEADER)
	pingMessage.data = pingMessage.header.decodeHeaderToBytes()

	_, err = c.Write(pingMessage.data)
	if err != nil {
		custError := "[FAILURE] Writing to connection failed:" + err.Error()
		log.Error(custError)
	}

	// receive KDM_PONG
	answer := readMessage(c)
	if answer == nil {
		return false
	}

	return answer.header.messageType == KDM_PONG

}

// finds k closest peers to given key
// if flag findValue ist set, then it searches for the stored value to the given key
func (thisNode *localNode) nodeLookup(key id, findValue bool) []peer {
	var closestPeersOld []peer

	waitingTime := 10
	for {
		if findValue {
			// if findValue is set, search in local hashTable
			_, ok := thisNode.hashTable.read(key)
			if ok {
				// value found, halt lookup process
				log.Debug("VALUE WAS FOUND IN LOCAL HASH TABLE OF ", Conf.apiPort)
				return nil
			}
		}

		// find k closest peers on local node
		closestPeersNew := thisNode.findNumberOfClosestPeersOnNode(key, Conf.k)
		if !wasAnyNewPeerAdded(closestPeersOld, closestPeersNew) {
			// if no new closer peer was found, check if waitingTime exceeds timeout
			if waitingTime > 1000 {
				break
			}
			// if not, increase waitingTime
			waitingTime = waitingTime * 10
		} else {
			// if new closer peer was found, reset waitingTime
			waitingTime = 10
		}

		// to every newly added close node, send KDM_FIND_VALUE or KDM_FIND_NODE (depending on boolean findValue)
		for _, p := range thisNode.findNumberOfClosestPeersOnNode(key, Conf.a) {
			if wasANewPeerAdded(closestPeersOld, p) {
				if findValue {
					msgBody := kdmFindValueBody{
						id: key,
					}
					m := makeP2PMessageOutOfBody(&msgBody, KDM_FIND_VALUE)
					sendP2PMessage(m, p)
				} else {
					msgBody := kdmFindNodeBody{
						id: key,
					}
					m := makeP2PMessageOutOfBody(&msgBody, KDM_FIND_NODE)
					sendP2PMessage(m, p)
				}
			}
		}
		closestPeersOld = closestPeersNew

		// give remote peers time to answer and give routing table time to update (at maximum ~1110 ms)
		time.Sleep(time.Duration(waitingTime) * time.Millisecond)
	}
	return closestPeersOld

}

// finds k closest nodes to given key on local node and generates body of KDM_FIND_NODE_ANSWER message
func (thisNode *localNode) FIND_NODE(key id) kdmFindNodeAnswerBody {

	closestPeers := thisNode.findNumberOfClosestPeersOnNode(key, Conf.k)
	answerBody := kdmFindNodeAnswerBody{answerPeers: closestPeers}
	log.Debug(answerBody)
	return answerBody
}

// locates k closest Nodes in network and sends KDM_STORE messages to them
func store(key id, value []byte, ttl uint16) {
	// locate k closest nodes in network
	kClosestPeers := thisNode.nodeLookup(key, false)
	log.Debug("FINAL : number of k CLOSEST PEERS", len(kClosestPeers))

	// send KDM_STORE messages to each of them
	for _, p := range kClosestPeers {
		storeBdy := kdmStoreBody{
			key:   key,
			value: value,
			ttl:   ttl,
		}
		m := makeP2PMessageOutOfBody(&storeBdy, KDM_STORE)
		sendP2PMessage(m, p)
	}
}

// checks every second if keys are expired or should be republished
func startTimers(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Debug("[DEBUG] Timer received stoppingSignal")
			return
		case <-ticker.C:
			// expire keys
			thisNode.hashTable.expireKeys()

			// republish keys
			thisNode.hashTable.republishKeys()
		}
	}
}
