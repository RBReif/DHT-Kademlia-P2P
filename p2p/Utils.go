package p2p

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
)

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

//returns maximum Size of the bucket at specified index
func maxSizeOfBucket(index int) int {
	max := k
	if index < k {
		rangeLimit := math.Pow(2, float64(index+1)) - math.Pow(2, float64(index))
		if rangeLimit < float64(max) {
			max = int(rangeLimit)
		}
	}

	return max
}

//returns index of the bucket that contains/is responsible for a given key
func (n *peer) findIndexOfResponsibleBucket(key id) int {
	d := distance(n.id, key)

	indexFirstRelevantByte := 19 //
	for i := 0; i < 20; i++ {
		if d[i] > 0 {
			indexFirstRelevantByte = i //this means we have 19-i trailing bytes after key[i]
			break
		}
	}
	i := 8*(19-indexFirstRelevantByte) + bits.Len8(d[indexFirstRelevantByte]) - 1
	fmt.Println("Bucket: ", i)
	return i
}

//returns index of the peer from a slice of peers that is the farest away from a given key
func findIndexOfFarestPeerInSlice(peers []peer, key id) int {
	var maxDistance [20]byte
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

//returns a number of the closest peers in a given bucket to a given key
func findNumberOfClosestPeersInOneBucket(kBucket []peer, key id, number int) []peer {
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

func (n *peer) findNumberOfClosestPeersOnNode(key id, number int) []peer {
	result := make([]peer, 0, number)
	indexOfResponsibleBucket := n.findIndexOfResponsibleBucket(key)
	result = append(result, findNumberOfClosestPeersInOneBucket(n.kBuckets[indexOfResponsibleBucket], key, number)...)
	if len(result) == number {
		return result
	}

	toBeFilled := number - len(result)
	tempResult := make([]peer, 0, 2*toBeFilled)
	for i := 0; i < 80; i++ {
		tempResult = append(tempResult, findNumberOfClosestPeersInOneBucket(n.kBuckets[(indexOfResponsibleBucket-i+160)%160], key, toBeFilled)...)
		tempResult = append(tempResult, findNumberOfClosestPeersInOneBucket(n.kBuckets[(indexOfResponsibleBucket+i)%160], key, toBeFilled)...)
		if len(tempResult) >= number {
			break
		}
		toBeFilled = number - len(result)
	}
	for len(tempResult) > number {
		indexOfFarest := findIndexOfFarestPeerInSlice(tempResult, key)
		tempResult = append(tempResult[:indexOfFarest], tempResult[indexOfFarest+1:]...)
	}
	result = append(result, tempResult...)
	return result
}

func wasNewPeerAdded(oldPeers []peer, newPeers []peer) bool {
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
