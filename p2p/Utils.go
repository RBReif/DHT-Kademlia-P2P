package p2p

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
