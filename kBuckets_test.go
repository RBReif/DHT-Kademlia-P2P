package main

import "testing"

func TestContains(t *testing.T) {
	kBucket := kBucket{}

	testPeer1 := peer{id: buildTestIdFromString("111")}
	testPeer2 := peer{id: buildTestIdFromString("110")}

	kBucket = append(kBucket, testPeer1)
	if !kBucket.contains(testPeer1.id) {
		t.Errorf("[FAILURE] id of peer not in k-Bucket even though it was inserted")
	}
	if kBucket.contains(testPeer2.id) {
		t.Errorf("[FAILURE] id of peer in k-Bucket even though it does not exist in k-Bucket")
	}
}

func TestIndexOf(t *testing.T) {
	kBucket := kBucket{}

	testPeer1 := peer{id: buildTestIdFromString("111")}
	testPeer2 := peer{id: buildTestIdFromString("110")}

	kBucket = append(kBucket, testPeer1)

	if kBucket.indexOf(testPeer1.id) != 0 {
		t.Errorf("[FAILURE] indexOf returned false index. Expected: 0")
	}
	if kBucket.indexOf(testPeer2.id) != -1 {
		t.Errorf("[FAILURE] indexOf returned false index. Expected: -1")
	}
}

func TestMoveToTail(t *testing.T) {
	kBucket := kBucket{}
	testPeer1 := peer{id: buildTestIdFromString("111")}
	testPeer2 := peer{id: buildTestIdFromString("110")}
	testPeer3 := peer{id: buildTestIdFromString("101")}
	testPeer4 := peer{id: buildTestIdFromString("100")}
	testPeer5 := peer{id: buildTestIdFromString("011")}
	testPeer6 := peer{id: buildTestIdFromString("010")}
	testPeer7 := peer{id: buildTestIdFromString("001")}

	kBucket = append(kBucket, testPeer1)
	kBucket = append(kBucket, testPeer2)
	kBucket = append(kBucket, testPeer3)
	kBucket = append(kBucket, testPeer4)
	kBucket = append(kBucket, testPeer5)
	kBucket = append(kBucket, testPeer6)
	kBucket = append(kBucket, testPeer7)

	if kBucket.indexOf(testPeer1.id) != 0 {
		t.Errorf("[FAILURE] indexOf returned false index. Expected: 0")
	}

	kBucket.moveToTail(testPeer1.id)
	if kBucket.indexOf(testPeer1.id) != 6 {
		t.Errorf("[FAILURE] moveToTail did not work")
	}

	// if id not existing, nothing should happen
	kBucketBefore := kBucket
	kBucket.moveToTail(buildTestIdFromString("111111111"))
	for i := range kBucketBefore {
		if kBucketBefore[i] != kBucket[i] {
			t.Errorf("[FAILURE] kBucket entries before and after invalid id not the same as expected")
		}
	}

}

func TestMaxSize(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	if routingTree.maxSize() != Conf.k {
		t.Errorf("[FAILURE] maxSize of root should be Conf.k")
	}

	tmpPrefix := ""

	// when prefix as long as SIZE_OF_ID * 8, then maxSize should be 2^0=1
	for i := 0; i < SIZE_OF_ID*8; i++ {
		tmpPrefix += "0"
	}
	routingTree.prefix = tmpPrefix[:len(tmpPrefix)]
	if routingTree.maxSize() != 1 {
		t.Errorf("[FAILURE] when prefix as long as SIZE_OF_ID * 8, then maxSize should be 1")
	}

	// when prefix as long as SIZE_OF_ID * 8 - 1, then maxSize should be 2^1=2
	routingTree.prefix = tmpPrefix[:len(tmpPrefix)-1]
	if routingTree.maxSize() != 2 {
		t.Errorf("[FAILURE] when prefix as long as SIZE_OF_ID * 8 - 1, then maxSize should be 2")
	}

	// when prefix as long as SIZE_OF_ID * 8 - 2, then maxSize should be 2^2=4
	routingTree.prefix = tmpPrefix[:len(tmpPrefix)-2]
	if routingTree.maxSize() != 4 {
		t.Errorf("[FAILURE] when prefix as long as SIZE_OF_ID * 8 - 2, then maxSize should be 4")
	}

	// when prefix as long as SIZE_OF_ID * 8 - 3, then maxSize should be 2^3=8 or in this case 5 (Conf.k)
	routingTree.prefix = tmpPrefix[:len(tmpPrefix)-3]
	if routingTree.maxSize() != 5 {
		t.Errorf("[FAILURE] when prefix as long as SIZE_OF_ID * 8 - 3, then maxSize should be 8 or in this case 5 (Conf.k)")
	}
}
func TestRemove(t *testing.T) {
	kBucket := kBucket{}
	testPeer1 := peer{id: buildTestIdFromString("111")}
	testPeer2 := peer{id: buildTestIdFromString("110")}
	testPeer3 := peer{id: buildTestIdFromString("101")}
	testPeer4 := peer{id: buildTestIdFromString("100")}
	testPeer5 := peer{id: buildTestIdFromString("011")}
	testPeer6 := peer{id: buildTestIdFromString("010")}
	testPeer7 := peer{id: buildTestIdFromString("001")}

	kBucket = append(kBucket, testPeer1)
	kBucket = append(kBucket, testPeer2)
	kBucket = append(kBucket, testPeer3)
	kBucket = append(kBucket, testPeer4)
	kBucket = append(kBucket, testPeer5)
	kBucket = append(kBucket, testPeer6)
	kBucket = append(kBucket, testPeer7)

	if kBucket.indexOf(testPeer1.id) != 0 {
		t.Errorf("[FAILURE] indexOf returned false index. Expected: 0")
	}

	kBucket.remove(testPeer1.id)
	if kBucket.indexOf(testPeer1.id) != -1 {
		t.Errorf("[FAILURE] indexOf returned false index. Expected: -1. Removing of id seems to not have worked correctly")
	}

}

func TestGetSibling(t *testing.T) {
	// init routingTree
	routingTable := routingTree{}
	routingTable.left = &routingTree{parent: &routingTable}
	routingTable.right = &routingTree{parent: &routingTable}
	routingTable.left.prefix = "0"
	routingTable.right.prefix = "1"

	// if current routingTree is left child of parent, getSibling should return right child of parent
	if routingTable.left != routingTable.right.getSibling() {
		t.Errorf("[FAILURE] getSibling() of right child has to return left child of parent")
	}

	// if current routingTree is right child of parent, getSibling should return left child of parent
	if routingTable.right != routingTable.left.getSibling() {
		t.Errorf("[FAILURE] getSibling() of left child has to return right child of parent")
	}

	// if current routingTree is root, getSibling should return nil
	if routingTable.getSibling() != nil {
		t.Errorf("[FAILURE] getSibling() of root has to return nil")
	}
}

func TestUpdateInsert(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	// init localNode
	thisNode := localNode{routingTree: *routingTree}
	testPeer1 := peer{id: buildTestIdFromString("111")}
	testPeer2 := peer{id: buildTestIdFromString("110")}
	testPeer3 := peer{id: buildTestIdFromString("101")}
	testPeer4 := peer{id: buildTestIdFromString("100")}
	testPeer5 := peer{id: buildTestIdFromString("011")}
	testPeer6 := peer{id: buildTestIdFromString("010")}
	testPeer7 := peer{id: buildTestIdFromString("001")}

	thisNode.thisPeer = peer{id: buildTestIdFromString("0")} // own id only 0s

	thisNode.updateRoutingTable(testPeer1)
	thisNode.updateRoutingTable(testPeer2)
	thisNode.updateRoutingTable(testPeer3)
	thisNode.updateRoutingTable(testPeer4)
	thisNode.updateRoutingTable(testPeer5)
	thisNode.updateRoutingTable(testPeer6)
	thisNode.updateRoutingTable(testPeer7)

	// TODO: implement correctness checks

}

func buildEmptyTestRoutingTree() *routingTree {
	result := routingTree{
		left:    nil,
		right:   nil,
		parent:  nil,
		prefix:  "",
		kBucket: kBucket{},
	}

	return &result
}

// takes a string of 0 and 1 and converts it to id
// chars other than 0 or 1 are replaced by 0s
// longer strings than id size are shortened and shorter strings are filled up with trailing 0s
func buildTestIdFromString(prefix string) id {
	if len(prefix) > SIZE_OF_ID*8 {
		prefix = prefix[:SIZE_OF_ID*8]
	}

	for len(prefix) < SIZE_OF_ID*8 {
		prefix += "0"
	}

	id := id{}
	for i := 0; i < SIZE_OF_ID; i++ {
		tmpByte := 0
		if prefix[i*8] == '1' {
			tmpByte = 1
		}

		for j := 1; j < 8; j++ {
			tmpByte = tmpByte << 1
			bit := 0
			if prefix[i*8+j] == '1' {
				bit = 1
			}
			tmpByte += bit
		}

		id[i] = byte(tmpByte)
	}

	return id

}
