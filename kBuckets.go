package main

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
)

type kBucket []peer

// struct for routing table binary tree
// if it has children, then kBucket has to be nil and vice versa
type routingTree struct {
	left    *routingTree
	right   *routingTree
	parent  *routingTree // nil if routingTree is root
	prefix  string
	kBucket kBucket
}

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

func (kBucket *routingTree) contains(id id) bool {

	for _, element := range kBucket.kBucket {
		if element.id == id {
			return true
		}
	}

	return false

}

func (kBucket *routingTree) indexOf(id id) int {

	for i, element := range kBucket.kBucket {
		if element.id == id {
			return i
		}
	}

	return -1

}

func (kBucket *routingTree) moveToTail(id id) {

	i := kBucket.indexOf(id)

	tmp := kBucket.kBucket[i]
	kBucket.kBucket = append(kBucket.kBucket[:i], kBucket.kBucket[i+1:]...)
	kBucket.kBucket = append(kBucket.kBucket, tmp)

}

func (kBucket *routingTree) isFull() bool {
	return len(kBucket.kBucket) == kBucket.maxSize()
}

func (kBucket *routingTree) maxSize() int {

	remainingBits := Conf.keySize - len(kBucket.prefix)
	if remainingBits < Conf.k { // roughly evict obvious cases
		rangeLimit := math.Pow(2, float64(remainingBits))
		if rangeLimit < float64(Conf.k) {
			return int(rangeLimit)
		}
	}

	return Conf.k

}

//returns maximum Size of the bucket at specified index
func maxSizeOfBucket(index int) int {
	max := Conf.k
	if index < Conf.k {
		rangeLimit := math.Pow(2, float64(index+1)) - math.Pow(2, float64(index))
		if rangeLimit < float64(max) {
			max = int(rangeLimit)
		}
	}

	return max
}

func (thisNode *localNode) findResponsibleBucket(key id) *routingTree {
	var tmpTree = thisNode.routingTree

	for {
		if tmpTree.kBucket != nil {
			return &tmpTree
		} else {
			var bitNumber = len(tmpTree.prefix) % 8
			var byteNumber = len(tmpTree.prefix) / 8
			var byte = key[byteNumber]
			var bit = (byte & (128 >> bitNumber)) != 0

			if bit == false {
				tmpTree = *tmpTree.left
			} else {
				tmpTree = *tmpTree.right
			}

		}
	}
}

func (kBucket *routingTree) insert(peer peer) {
	// TODO: only insert if not already in bucket
	// only insert if not already full
	if len(kBucket.kBucket) < kBucket.maxSize() {
		kBucket.kBucket = append(kBucket.kBucket, peer)
	}
}

func (kBucket *routingTree) remove(id id) {
	i := kBucket.indexOf(id)
	kBucket.kBucket = append(kBucket.kBucket[:i], kBucket.kBucket[i+1:]...)
}

func (kBucket *routingTree) split() {
	prefixLeft := kBucket.prefix + "0"
	prefixRight := kBucket.prefix + "1"

	kBucketLeft := routingTree{prefix: prefixLeft}
	kBucketRight := routingTree{prefix: prefixRight}

	kBucket.left = &kBucketLeft
	kBucket.right = &kBucketRight

	for _, element := range kBucket.kBucket {
		if element.id.startsWith(prefixLeft) {
			kBucket.left.insert(element)
		} else {
			kBucket.right.insert(element)
		}
	}

	kBucket.kBucket = nil

}

// returns the corresponding sibling (left child if routingTree is right child and vice versa)
// or nil if current routingTree is root
func (kBucket *routingTree) getSibling() *routingTree {
	if kBucket.parent == nil {
		return nil
	}
	if kBucket.parent.left == kBucket {
		return kBucket.parent.right
	} else {
		return kBucket.parent.left
	}
}

func (kBucket *routingTree) getNumberOfClosestPeers(key id, number int) []peer {
	if kBucket.kBucket != nil {
		return findNumberOfClosestPeersInOneBucket(kBucket.kBucket, key, number)
	} else {
		tmp := kBucket.left.getNumberOfClosestPeers(key, number)
		tmp = append(tmp, kBucket.right.getNumberOfClosestPeers(key, number)...)
		return findNumberOfClosestPeersInOneBucket(tmp, key, number)
	}
}

//returns index of the bucket that contains/is responsible for a given id
func (thisNode *localNode) findIndexOfResponsibleBucket(key id) int {
	d := distance(thisNode.thisPeer.id, key)

	indexFirstRelevantByte := 19 // [0 0 0 0 0 0 0 ... 0 0 0 9 3 23]
	for i := 0; i < 20; i++ {
		if d[i] > 0 {
			indexFirstRelevantByte = i //this means we have 19-i trailing bytes after id[i]
			break
		}
	}
	i := 8*(19-indexFirstRelevantByte) + bits.Len8(d[indexFirstRelevantByte]) - 1
	fmt.Println("Bucket: ", i)
	return i
}

//returns index of the peer from a slice of peers that is the farest away from a given id
func findIndexOfFarestPeerInSlice(peers []peer, key id) int {
	var maxDistance [SIZE_OF_ID]byte
	index := -1
	for i := 0; i < len(peers); i++ {
		d := distance(key, peers[i].id)
		if bytes.Compare(d[:], maxDistance[:]) > 0 {
			maxDistance = d
			index = i
		}
	}
	return index
}

//returns a number of the closest peers in a given bucket to a given id
func findNumberOfClosestPeersInOneBucket(kBucket kBucket, key id, number int) []peer {
	result := make([]peer, 0, number)

	for i := 0; i < len(kBucket); i++ {
		if len(result) < number {
			result = append(result, kBucket[i])
		} else {
			indexOfFarest := findIndexOfFarestPeerInSlice(result, key)
			dNew := distance(key, kBucket[i].id)
			dOld := distance(key, result[indexOfFarest].id)
			if bytes.Compare(dNew[:], dOld[:]) < 0 {
				result[indexOfFarest] = kBucket[i]
			}
		}
	}
	return result
}

//returns a specified amount of peers that are the closest to a specified id on a node
func (thisNode *localNode) findNumberOfClosestPeersOnNode(key id, number int) []peer {
	result := make([]peer, 0, number)
	responsibleBucket := thisNode.findResponsibleBucket(key)

	for {
		result = responsibleBucket.getNumberOfClosestPeers(key, number)
		if len(result) == number || responsibleBucket.parent == nil {
			return result
		} else {
			responsibleBucket = responsibleBucket.parent
		}
	}
}

//returns whether there was any new peer added in newPeers that has not been there before in oldPeers
func wasAnyNewPeerAdded(oldPeers []peer, newPeers []peer) bool {
	for i := 0; i < len(newPeers); i++ {
		isThere := false
		for j := 0; j < len(oldPeers); j++ {
			if bytes.Compare(newPeers[i].id[:], oldPeers[j].id[:]) == 0 {
				isThere = true
				break
			}
		}
		if !isThere {
			return true
		}

	}
	return false
}

func wasANewPeerAdded(oldPeers []peer, newPeer peer) bool {
	for j := 0; j < len(oldPeers); j++ {
		if bytes.Compare(newPeer.id[:], oldPeers[j].id[:]) == 0 {
			return false
		}
	}
	return true
}

func (thisNode *localNode) updateKBucket(p peer) {
	// find responsible k-Bucket
	kBucket := thisNode.findResponsibleBucket(p.id)

	// if peer already exists in k-Bucket, move it to the tail of the list
	if kBucket.contains(p.id) {
		kBucket.moveToTail(p.id)
	} else { // else
		if !kBucket.isFull() {
			// if k-Bucket is not already full, insert peer
			kBucket.insert(p)
		} else { // if k-Bucket is already full
			// if k-Bucket includes own id, split bucket and repeat insertion attempt
			if kBucket.contains(thisNode.thisPeer.id) {
				kBucket.split()
				thisNode.updateKBucket(p)
			} else {
				// else ping least-recently seen node
				nodeActive := pingNode(kBucket.kBucket[0])

				if !nodeActive {
					// if node is inactive, discard least-recently seen node and insert the new peer at the tail
					kBucket.remove(kBucket.kBucket[0].id)
					kBucket.insert(p)

				} else {
					// if node is active, discard peer and move least-recently seen node to the tail
					kBucket.moveToTail(kBucket.kBucket[0].id)
				}
			}
		}
	}

}
