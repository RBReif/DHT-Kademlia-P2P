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

func TestSplit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("[FAILURE] split() did not panic even though it exceeded key-length")
		}
	}()

	thisNode := localNode{thisPeer: peer{id: buildTestIdFromString("")}, routingTree: routingTree{kBucket: kBucket{}}}
	thisNode.updateRoutingTable(thisNode.thisPeer)
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

func TestFindNumberOfClosestPeersInOneBucket(t *testing.T) {
	// should return empty slice if bucket is empty
	testBucket := kBucket{}
	if len(testBucket.findNumberOfClosestPeersInOneBucket(buildTestIdFromString(""), 1)) != 0 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return empty slice if bucket is empty")
	}

	// test with plausible values
	testPeer1 := peer{id: buildTestIdFromString("0001")}
	testPeer2 := peer{id: buildTestIdFromString("0010")}
	testPeer3 := peer{id: buildTestIdFromString("0011")}
	testPeer4 := peer{id: buildTestIdFromString("0100")}
	testPeer5 := peer{id: buildTestIdFromString("0101")}
	testPeer6 := peer{id: buildTestIdFromString("0110")}
	testPeer7 := peer{id: buildTestIdFromString("0111")}
	testPeer8 := peer{id: buildTestIdFromString("1000")}
	testPeer9 := peer{id: buildTestIdFromString("1001")}
	testPeer10 := peer{id: buildTestIdFromString("1010")}
	testPeer11 := peer{id: buildTestIdFromString("1011")}
	testPeer12 := peer{id: buildTestIdFromString("1100")}
	testPeer13 := peer{id: buildTestIdFromString("1101")}
	testPeer14 := peer{id: buildTestIdFromString("1110")}
	testPeer15 := peer{id: buildTestIdFromString("1111")}

	testBucket = append(testBucket, testPeer1)
	testBucket = append(testBucket, testPeer2)
	testBucket = append(testBucket, testPeer3)
	testBucket = append(testBucket, testPeer4)
	testBucket = append(testBucket, testPeer5)
	testBucket = append(testBucket, testPeer6)
	testBucket = append(testBucket, testPeer7)
	testBucket = append(testBucket, testPeer8)
	testBucket = append(testBucket, testPeer9)
	testBucket = append(testBucket, testPeer10)
	testBucket = append(testBucket, testPeer11)
	testBucket = append(testBucket, testPeer12)
	testBucket = append(testBucket, testPeer13)
	testBucket = append(testBucket, testPeer14)
	testBucket = append(testBucket, testPeer15)

	result := testBucket.findNumberOfClosestPeersInOneBucket(testPeer8.id, 3)
	if len(result) != 3 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return 3 peers in this case")
	}

	if result[0] != testPeer8 {
		t.Errorf("[FAILURE] closest peer of testPeer8 should be testPeer8")
	}

	if result[1] != testPeer9 {
		t.Errorf("[FAILURE] second closest peer of testPeer8 should be testPeer9")
	}

	if result[2] != testPeer10 {
		t.Errorf("[FAILURE] third closest peer of testPeer8 should be testPeer10")
	}
}

func TestGetNumberOfClosestPeersOnNode(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	// init localNode
	thisNode := localNode{routingTree: *routingTree}
	testPeer1 := peer{id: buildTestIdFromString("0001")}
	testPeer2 := peer{id: buildTestIdFromString("0010")}
	testPeer3 := peer{id: buildTestIdFromString("0011")}
	testPeer4 := peer{id: buildTestIdFromString("0100")}
	testPeer5 := peer{id: buildTestIdFromString("0101")}
	testPeer6 := peer{id: buildTestIdFromString("0110")}
	testPeer7 := peer{id: buildTestIdFromString("0111")}
	testPeer8 := peer{id: buildTestIdFromString("1000")}
	testPeer9 := peer{id: buildTestIdFromString("1001")}
	testPeer10 := peer{id: buildTestIdFromString("1010")}
	testPeer11 := peer{id: buildTestIdFromString("1011")}
	testPeer12 := peer{id: buildTestIdFromString("1100")}
	testPeer13 := peer{id: buildTestIdFromString("1101")}
	testPeer14 := peer{id: buildTestIdFromString("1110")}
	testPeer15 := peer{id: buildTestIdFromString("1111")}

	thisNode.thisPeer = peer{id: buildTestIdFromString("0")} // own id only 0s

	thisNode.updateRoutingTable(testPeer1)
	thisNode.updateRoutingTable(testPeer2)
	thisNode.updateRoutingTable(testPeer3)
	thisNode.updateRoutingTable(testPeer4)
	thisNode.updateRoutingTable(testPeer5)
	thisNode.updateRoutingTable(testPeer6)
	thisNode.updateRoutingTable(testPeer7)
	thisNode.updateRoutingTable(testPeer8)
	thisNode.updateRoutingTable(testPeer9)
	thisNode.updateRoutingTable(testPeer10)
	thisNode.updateRoutingTable(testPeer11)
	thisNode.updateRoutingTable(testPeer12)
	thisNode.updateRoutingTable(testPeer13)
	thisNode.updateRoutingTable(testPeer14)
	thisNode.updateRoutingTable(testPeer15)

	// findNumberOfClosestPeersOnNode(...) searching for own id (left subtree in this case) should return the closest peers existing
	result := thisNode.findNumberOfClosestPeersOnNode(thisNode.thisPeer.id, 3)
	if len(result) != 3 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return 3 peers in this case")
	}

	if result[0] != testPeer1 {
		t.Errorf("[FAILURE] closest peer of thisPeer should be testPeer1")
	}

	if result[1] != testPeer2 {
		t.Errorf("[FAILURE] second closest peer of thisPeer should be testPeer2")
	}

	if result[2] != testPeer3 {
		t.Errorf("[FAILURE] third closest peer of thisPeer should be testPeer3")
	}

	// getNumberOfClosestPeers(...) of root routingTree searching for id in right subtree should return the closest peers of the last seen peers
	result = thisNode.findNumberOfClosestPeersOnNode(testPeer8.id, 3)
	if len(result) != 3 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return 3 peers in this case")
	}

	if result[0] != testPeer11 {
		t.Errorf("[FAILURE] closest peer of testPeer8 should be testPeer8")
	}

	if result[1] != testPeer12 {
		t.Errorf("[FAILURE] second closest peer of testPeer8 should be testPeer9")
	}

	if result[2] != testPeer13 {
		t.Errorf("[FAILURE] third closest peer of testPeer8 should be testPeer10")
	}

}

func TestGetNumberOfClosestPeers(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	// init localNode
	thisNode := localNode{routingTree: *routingTree}
	testPeer1 := peer{id: buildTestIdFromString("0001")}
	testPeer2 := peer{id: buildTestIdFromString("0010")}
	testPeer3 := peer{id: buildTestIdFromString("0011")}
	testPeer4 := peer{id: buildTestIdFromString("0100")}
	testPeer5 := peer{id: buildTestIdFromString("0101")}
	testPeer6 := peer{id: buildTestIdFromString("0110")}
	testPeer7 := peer{id: buildTestIdFromString("0111")}
	testPeer8 := peer{id: buildTestIdFromString("1000")}
	testPeer9 := peer{id: buildTestIdFromString("1001")}
	testPeer10 := peer{id: buildTestIdFromString("1010")}
	testPeer11 := peer{id: buildTestIdFromString("1011")}
	testPeer12 := peer{id: buildTestIdFromString("1100")}
	testPeer13 := peer{id: buildTestIdFromString("1101")}
	testPeer14 := peer{id: buildTestIdFromString("1110")}
	testPeer15 := peer{id: buildTestIdFromString("1111")}

	thisNode.thisPeer = peer{id: buildTestIdFromString("0")} // own id only 0s

	thisNode.updateRoutingTable(testPeer1)
	thisNode.updateRoutingTable(testPeer2)
	thisNode.updateRoutingTable(testPeer3)
	thisNode.updateRoutingTable(testPeer4)
	thisNode.updateRoutingTable(testPeer5)
	thisNode.updateRoutingTable(testPeer6)
	thisNode.updateRoutingTable(testPeer7)
	thisNode.updateRoutingTable(testPeer8)
	thisNode.updateRoutingTable(testPeer9)
	thisNode.updateRoutingTable(testPeer10)
	thisNode.updateRoutingTable(testPeer11)
	thisNode.updateRoutingTable(testPeer12)
	thisNode.updateRoutingTable(testPeer13)
	thisNode.updateRoutingTable(testPeer14)
	thisNode.updateRoutingTable(testPeer15)

	// getNumberOfClosestPeers(...) of root routingTree searching for own id (left subtree in this case) should return the closest peers existing
	result := thisNode.routingTree.getNumberOfClosestPeers(thisNode.thisPeer.id, 3)
	if len(result) != 3 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return 3 peers in this case")
	}

	if result[0] != testPeer1 {
		t.Errorf("[FAILURE] closest peer of thisPeer should be testPeer1")
	}

	if result[1] != testPeer2 {
		t.Errorf("[FAILURE] second closest peer of thisPeer should be testPeer2")
	}

	if result[2] != testPeer3 {
		t.Errorf("[FAILURE] third closest peer of thisPeer should be testPeer3")
	}

	// getNumberOfClosestPeers(...) of root routingTree searching for id in right subtree should return the closest peers of the last seen peers
	result = thisNode.routingTree.getNumberOfClosestPeers(testPeer8.id, 3)
	if len(result) != 3 {
		t.Errorf("[FAILURE] findNumberOfClosestPeersInOneBucket(...) should return 3 peers in this case")
	}

	if result[0] != testPeer11 {
		t.Errorf("[FAILURE] closest peer of testPeer8 should be testPeer8")
	}

	if result[1] != testPeer12 {
		t.Errorf("[FAILURE] second closest peer of testPeer8 should be testPeer9")
	}

	if result[2] != testPeer13 {
		t.Errorf("[FAILURE] third closest peer of testPeer8 should be testPeer10")
	}

}

func TestUpdateInsertAndSplit(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	// init localNode
	thisNode := localNode{routingTree: *routingTree}
	testPeer1 := peer{id: buildTestIdFromString("0001")}
	testPeer2 := peer{id: buildTestIdFromString("0010")}
	testPeer3 := peer{id: buildTestIdFromString("0011")}
	testPeer4 := peer{id: buildTestIdFromString("0100")}
	testPeer5 := peer{id: buildTestIdFromString("0101")}
	testPeer6 := peer{id: buildTestIdFromString("0110")}
	testPeer7 := peer{id: buildTestIdFromString("0111")}
	testPeer8 := peer{id: buildTestIdFromString("1000")}
	testPeer9 := peer{id: buildTestIdFromString("1001")}
	testPeer10 := peer{id: buildTestIdFromString("1010")}
	testPeer11 := peer{id: buildTestIdFromString("1011")}
	testPeer12 := peer{id: buildTestIdFromString("1100")}
	testPeer13 := peer{id: buildTestIdFromString("1101")}
	testPeer14 := peer{id: buildTestIdFromString("1110")}
	testPeer15 := peer{id: buildTestIdFromString("1111")}

	thisNode.thisPeer = peer{id: buildTestIdFromString("0")} // own id only 0s

	thisNode.updateRoutingTable(testPeer1)
	thisNode.updateRoutingTable(testPeer2)
	thisNode.updateRoutingTable(testPeer3)
	thisNode.updateRoutingTable(testPeer4)
	thisNode.updateRoutingTable(testPeer5)
	thisNode.updateRoutingTable(testPeer6)
	thisNode.updateRoutingTable(testPeer7)
	thisNode.updateRoutingTable(testPeer8)
	thisNode.updateRoutingTable(testPeer9)
	thisNode.updateRoutingTable(testPeer10)
	thisNode.updateRoutingTable(testPeer11)
	thisNode.updateRoutingTable(testPeer12)
	thisNode.updateRoutingTable(testPeer13)
	thisNode.updateRoutingTable(testPeer14)
	thisNode.updateRoutingTable(testPeer15)

	// TODO: implement correctness checks

}

func TestUpdateInsertAndPing(t *testing.T) {
	// init Conf
	Conf.k = 5

	// init empty routingTree
	routingTree := buildEmptyTestRoutingTree()

	// init localNode
	thisNode := localNode{routingTree: *routingTree}
	testPeer1 := peer{id: buildTestIdFromString("0001")}
	testPeer2 := peer{id: buildTestIdFromString("0010")}
	testPeer3 := peer{id: buildTestIdFromString("0011")}
	testPeer4 := peer{id: buildTestIdFromString("0100")}
	testPeer5 := peer{id: buildTestIdFromString("0101")}
	testPeer6 := peer{id: buildTestIdFromString("0110")}
	testPeer7 := peer{id: buildTestIdFromString("0111")}
	testPeer8 := peer{id: buildTestIdFromString("1000")}
	testPeer9 := peer{id: buildTestIdFromString("1001")}
	testPeer10 := peer{id: buildTestIdFromString("1010")}
	testPeer11 := peer{id: buildTestIdFromString("1011")}
	testPeer12 := peer{id: buildTestIdFromString("1100")}
	testPeer13 := peer{id: buildTestIdFromString("1101")}
	testPeer14 := peer{id: buildTestIdFromString("1110")}
	testPeer15 := peer{id: buildTestIdFromString("1111")}

	thisNode.thisPeer = peer{id: buildTestIdFromString("0")} // own id only 0s

	thisNode.updateRoutingTable(testPeer1)
	thisNode.updateRoutingTable(testPeer2)
	thisNode.updateRoutingTable(testPeer3)
	thisNode.updateRoutingTable(testPeer4)
	thisNode.updateRoutingTable(testPeer5)
	thisNode.updateRoutingTable(testPeer6)
	thisNode.updateRoutingTable(testPeer7)
	thisNode.updateRoutingTable(testPeer8)
	thisNode.updateRoutingTable(testPeer9)
	thisNode.updateRoutingTable(testPeer10)
	thisNode.updateRoutingTable(testPeer11)
	thisNode.updateRoutingTable(testPeer12)
	thisNode.updateRoutingTable(testPeer13)
	thisNode.updateRoutingTable(testPeer14)
	thisNode.updateRoutingTable(testPeer15)

	// make one of the peers active
	testPeer9.ip = "127.0.0.1"
	testPeer9.port = 8000

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
