package p2p

// TODO: move k to singleton or something
var k int

type id [20]byte

type peer struct {
	ip   string
	port int
	id   id
}

// TODO: find better concept for node/peer
type node struct {
	peer    peer
	kBucket []peer
}

// TODO: move to dedicated file message.go
type message struct {
	sender peer
	// TODO
}

func (node *node) onMessageReceived(message message) {

	// if id of sender exists in kBucket, move to tail of kBucket
	if index, inBucket := isIdInKBucket(node.kBucket, message.sender.id); inBucket {
		moveToTail(node.kBucket, index)
	} else {
		// if kBucket has fewer than k entries, insert id to kBucket
		if len(node.kBucket) < k {
			node.kBucket = append(node.kBucket, message.sender)
		} else {
			// else ping least-recently seen node
			// TODO

			// if node not responding, remove node and insert the new one
			node.kBucket = append(node.kBucket[:index], node.kBucket[index+1:]...)
			node.kBucket = append(node.kBucket, message.sender)

			// else move node to tail and discard the new one
			moveToTail(node.kBucket, index)
		}

	}

}

// TODO: move everything below to utils

func moveToTail(kBucket []peer, i int) []peer {

	tmp := kBucket[i]
	kBucket = append(kBucket[:i], kBucket[i+1:]...)
	kBucket = append(kBucket, tmp)

	return kBucket

}

func isIdInKBucket(kBucket []peer, id id) (int, bool) {

	for index, element := range kBucket {
		if element.id == id {
			return index, true
		}
	}

	return -1, false

}

// distance function of kademlia
func distance(id1 id, id2 id) id {

	xor := make(id)

	for i := 0; i < len(id1); i++ {

		xor[i] = id1[i] ^ id2[i]

	}

	return xor

}
