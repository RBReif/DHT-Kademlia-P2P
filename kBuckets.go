package main

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"math"
	"sort"
	"strconv"
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

func (r *routingTree) toString() string {
	result := r.prefix
	if r.kBucket != nil {
		result = result + " has " + r.kBucket.toString() + "\n"
	}
	if r.left != nil {
		result = result + "left:" + r.left.toString() + "\n"
	}
	if r.right != nil {

		result = result + "right:" + r.right.toString() + "\n"
	}
	return result
}

// returns if kBucket contains id
func (kBucket *kBucket) contains(id id) bool {

	for _, element := range *kBucket {
		if element.id == id {
			return true
		}
	}

	return false

}

func (kBucket *kBucket) toString() string {
	result := ""
	for _, element := range *kBucket {
		result = result + "    " + strconv.Itoa(int(element.port))
	}

	return result

}

// checks if k-Bucket of given routingTree node is responsible for given id
func (routingTable *routingTree) inRange(id id) bool {
	return id.startsWith(routingTable.prefix)
}

// returns index of id in k-Bucket or -1 if not containing
func (kBucket *kBucket) indexOf(id id) int {

	for i, element := range *kBucket {
		if element.id == id {
			return i
		}
	}

	return -1

}

// moves id stored in k-Bucket to tail of k-Bucket
func (kBucket *kBucket) moveToTail(id id) {

	i := kBucket.indexOf(id)

	if i != -1 {
		tmp := (*kBucket)[i]
		*kBucket = append((*kBucket)[:i], (*kBucket)[i+1:]...)
		*kBucket = append(*kBucket, tmp)
	}

}

// checks if k-Bucket of given routingTree node is full
func (routingTable *routingTree) isFull() bool {
	return len(routingTable.kBucket) == routingTable.maxSize()
}

// returns maximum size of k-Bucket of given routingTree node
func (routingTable *routingTree) maxSize() int {

	remainingBits := SIZE_OF_ID*8 - len(routingTable.prefix)
	if remainingBits < Conf.k { // roughly evict obvious cases
		rangeLimit := math.Pow(2, float64(remainingBits))
		if rangeLimit < float64(Conf.k) {
			return int(rangeLimit)
		}
	}

	return Conf.k

}

// finds responsible routingTree node for given id
func (thisNode *localNode) findResponsibleRoutingTree(key id) *routingTree {
	var tmpTree = &thisNode.routingTree

	for {
		if tmpTree.kBucket != nil {
			return tmpTree
		} else {
			var bitNumber = len(tmpTree.prefix) % 8
			var byteNumber = len(tmpTree.prefix) / 8
			var byte = key[byteNumber]
			var bit = (byte & (128 >> bitNumber)) != 0

			if !bit {
				tmpTree = tmpTree.left
			} else {
				tmpTree = tmpTree.right
			}

		}
	}
}

// inserts new peer into k-Bucket of given routingTree node
func (routingTable *routingTree) insert(peer peer) {
	// only insert if id of peer is not already in bucket and if bucket is not already full
	if !routingTable.kBucket.contains(peer.id) && len(routingTable.kBucket) < routingTable.maxSize() {
		routingTable.kBucket = append(routingTable.kBucket, peer)
	}
}

// removes id from given k-Bucket
func (kBucket *kBucket) remove(id id) {
	i := kBucket.indexOf(id)
	*kBucket = append((*kBucket)[:i], (*kBucket)[i+1:]...)
}

// splits given routingTree node into two new routingTree nodes
func (routingTable *routingTree) split() error {

	if len(routingTable.prefix) == SIZE_OF_ID*8 {
		return errors.New("Tried to split k-Bucket with maximum size of 1")
	}

	log.Debug(thisNode.thisPeer.port, ": splitting at current prefix ", routingTable.prefix)
	prefixLeft := routingTable.prefix + "0"
	prefixRight := routingTable.prefix + "1"

	routingTreeLeft := routingTree{prefix: prefixLeft, parent: routingTable, kBucket: kBucket{}}
	routingTreeRight := routingTree{prefix: prefixRight, parent: routingTable, kBucket: kBucket{}}

	routingTable.left = &routingTreeLeft
	routingTable.right = &routingTreeRight

	for _, element := range routingTable.kBucket {
		if element.id.startsWith(prefixLeft) {
			routingTable.left.insert(element)
		} else {
			routingTable.right.insert(element)
		}
	}

	// reset kBucket of routingTree node
	routingTable.kBucket = nil

	return nil

}

// returns the corresponding sibling (left child if routingTree is right child and vice versa)
// or nil if current routingTree is root
func (routingTable *routingTree) getSibling() *routingTree {
	if routingTable.parent == nil {
		return nil
	}
	if routingTable.parent.left == routingTable {
		return routingTable.parent.right
	} else {
		return routingTable.parent.left
	}
}

//returns a number of the closest peers in a given bucket to a given id ordered by distance (closest at the beginning)
func (kBucket *kBucket) findNumberOfClosestPeersInOneBucket(key id, number int) []peer {
	result := make([]peer, 0, number)

	for i := 0; i < len(*kBucket); i++ {
		if len(result) < number {
			result = append(result, (*kBucket)[i])
		} else {
			indexOfFarest := findIndexOfFarestPeerInSlice(result, key)
			dNew := distance(key, (*kBucket)[i].id)
			dOld := distance(key, result[indexOfFarest].id)
			if bytes.Compare(dNew[:], dOld[:]) < 0 {
				result[indexOfFarest] = (*kBucket)[i]
			}
		}
	}

	// sort by distance
	sort.Slice(result, func(i, j int) bool {
		dI := distance(key, result[i].id)
		dJ := distance(key, result[j].id)
		return bytes.Compare(dI[:], dJ[:]) < 0
	})
	return result
}

// returns given number of closest peers to given id available in subtree with given routingTree node as root node
func (routingTable *routingTree) getNumberOfClosestPeers(key id, number int) []peer {
	if routingTable.kBucket != nil {
		return routingTable.kBucket.findNumberOfClosestPeersInOneBucket(key, number)
	} else {
		tmp := kBucket{}
		tmp = append(tmp, routingTable.left.getNumberOfClosestPeers(key, number)...)
		tmp = append(tmp, routingTable.right.getNumberOfClosestPeers(key, number)...)
		return tmp.findNumberOfClosestPeersInOneBucket(key, number)
	}
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

//returns a specified amount of peers that are the closest to a specified id on a node
func (thisNode *localNode) findNumberOfClosestPeersOnNode(key id, number int) []peer {
	responsibleBucket := thisNode.findResponsibleRoutingTree(key)

	for {
		result := responsibleBucket.getNumberOfClosestPeers(key, number)
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
			if bytes.Equal(newPeers[i].id[:], oldPeers[j].id[:]) {
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

// checks if given peer was added to given slice of peers
func wasANewPeerAdded(oldPeers []peer, newPeer peer) bool {
	for j := 0; j < len(oldPeers); j++ {
		if bytes.Equal(newPeer.id[:], oldPeers[j].id[:]) {
			return false
		}
	}
	return true
}

// updates routingTable of local node when it made contact with given peer
func (thisNode *localNode) updateRoutingTable(p peer) {
	// find responsible k-Bucket
	routingTree := thisNode.findResponsibleRoutingTree(p.id)

	// if peer already exists in k-Bucket, move it to the tail of the list
	if routingTree.kBucket.contains(p.id) {
		routingTree.kBucket.moveToTail(p.id)
	} else { // else
		if !routingTree.isFull() {
			// if k-Bucket is not already full, insert peer

			routingTree.insert(p)
		} else { // if k-Bucket is already full
			// if range of k-Bucket includes own id, split bucket and repeat insertion attempt
			if routingTree.inRange(thisNode.thisPeer.id) {
				err := routingTree.split()
				if err != nil {
					return // abort update process
				}
				thisNode.updateRoutingTable(p)
			} else {
				// else ping least-recently seen node
				nodeActive := pingNode(routingTree.kBucket[0], thisNode.thisPeer)

				if !nodeActive {
					// if node is inactive, discard least-recently seen node and insert the new peer at the tail
					routingTree.kBucket.remove(routingTree.kBucket[0].id)

					routingTree.insert(p)

				} else {
					// if node is active, discard peer and move least-recently seen node to the tail
					routingTree.kBucket.moveToTail(routingTree.kBucket[0].id)
				}
			}
		}
	}

}
