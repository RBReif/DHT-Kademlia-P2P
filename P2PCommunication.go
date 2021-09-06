package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"strings"

	//"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

var thisNode localNode

type localNode struct {
	thisPeer peer
	kBuckets [160][]peer
}

type peer struct {
	ip string
	//isIpv4 bool
	port uint16
	id   id
}

func (p *peer) toString() string {
	return "" + p.ip + ":" + strconv.Itoa(int(p.port)) + " (" + bytesToString(p.id.toByte()) + ")"
}

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

func initializeP2Pcomm() {
	//first we read the private key
	priv, err := ioutil.ReadFile(Conf.HostKeyFile)
	if err != nil {
		fmt.Println("Error while reading File: ", err)
	}
	fmt.Println("Bytes from private key pem file: ", priv)
	block, _ := pem.Decode([]byte(priv))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("failed to decode PEM block containing public key")
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
	fmt.Println("public key generated: ", pubKeyPem)

	fmt.Println("Public Key as bytes: ", publicKeyDer)

	//now we calculate the sha256 hash sum to retreive our ID
	h := sha256.New()
	h.Write(publicKeyDer)
	newID := h.Sum(nil)
	fmt.Println("Our ID: ", newID)

}

func extractPeer(line string) peer {
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
		fmt.Println("Wrong configuration: the port of ", line, " is not an Integer")
		os.Exit(1)
	}
	result.ip = ip
	result.port = uint16(port)
	return result
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
		hashTable[m.body.(*kdmStoreBody).key] = m.body.(*kdmStoreBody).value
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
