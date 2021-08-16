package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	ran "math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func makeApiMessageOutOfBody(msgBody apiBody, msgType uint16) apiMessage {
	//building header
	hdr := apiHeader{}
	msg := apiMessage{
		header: hdr,
	}
	switch msgType {
	case dhtSUCCESS:
		msg.header.size = uint16(2 + 2 + len(msgBody.(*successBody).key) + len(msgBody.(*successBody).value))
		msg.header.messageType = dhtSUCCESS
		msg.body = msgBody
	case dhtFAILURE:
		msg.header.size = uint16(2 + 2 + len(msgBody.(*failureBody).key))
		msg.header.messageType = dhtFAILURE
		msg.body = msgBody
	case dhtGET:
		msg.header.size = uint16(2 + 2 + len(msgBody.(*getBody).key))
		msg.header.messageType = dhtGET
		msg.body = msgBody
	case dhtPUT:
		msg.header.size = uint16(2 + 2 + 2 + 1 + 1 + len(msgBody.(*putBody).key) + len(msgBody.(*putBody).value))
		msg.header.messageType = dhtPUT
		msg.body = msgBody
	}
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[:2], msg.header.size)
	binary.BigEndian.PutUint16(data[2:4], msg.header.messageType)
	data = append(data, msg.body.decodeBodyToBytes()...)
	msg.data = data
	return msg
}

func TestGetCodingAndDecoding(t *testing.T) {

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	getBdy := getBody{key: i}

	get1 := makeApiMessageOutOfBody(&getBdy, dhtGET)
	fmt.Println("Get_1: ", get1.toString())
	if get1.body == nil {
		t.Errorf("Body of Get message is  nil")
	}

	fmt.Println(get1.data)

	get2 := makeApiMessageOutOfBytes(get1.data)
	fmt.Println("get2: ", get2.toString())

	if get1.header.size != get2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if get1.header.messageType != get2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(get1.body, get2.body) {
		t.Errorf("Parsing of Body (get)  does not work")
	}

}

func TestFailureCodingAndDecoding(t *testing.T) {

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	failureBdy := failureBody{key: i}

	failure1 := makeApiMessageOutOfBody(&failureBdy, dhtFAILURE)
	fmt.Println("Failure1: ", failure1.toString())
	if failure1.body == nil {
		t.Errorf("Body of Failure message is  nil")
	}

	fmt.Println(failure1.data)

	failure2 := makeApiMessageOutOfBytes(failure1.data)
	fmt.Println("failure2: ", failure2.toString())

	if failure1.header.size != failure2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if failure1.header.messageType != failure2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(failure1.body, failure2.body) {
		t.Errorf("Parsing of Body (failure)  does not work")
	}

}

func TestSuccessCodingAndDecoding(t *testing.T) {

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	successBdy := successBody{
		key:   i,
		value: value,
	}

	success1 := makeApiMessageOutOfBody(&successBdy, dhtSUCCESS)
	fmt.Println("Success1: ", success1.toString())
	if success1.body == nil {
		t.Errorf("Body of Success message is  nil")
	}

	fmt.Println(success1.data)

	success2 := makeApiMessageOutOfBytes(success1.data)
	fmt.Println("success2: ", success2.toString())

	if success1.header.size != success2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if success1.header.messageType != success2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(success1.body, success2.body) {
		t.Errorf("Parsing of Body (success)  does not work")
	}

}

func TestPutCodingAndDecoding(t *testing.T) {

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	putBdy := putBody{
		ttl:         20,
		reserved:    2,
		replication: 3,
		key:         i,
		value:       value,
	}

	put1 := makeApiMessageOutOfBody(&putBdy, dhtPUT)
	fmt.Println("put1: ", put1.toString())
	if put1.body == nil {
		t.Errorf("Body of Put message is  nil")
	}

	fmt.Println(put1.data)

	put2 := makeApiMessageOutOfBytes(put1.data)
	fmt.Println("put2: ", put2.toString())

	if put1.header.size != put2.header.size {
		t.Errorf("Parsing of Header size does not work")
	}
	if put1.header.messageType != put2.header.messageType {
		t.Errorf("Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(put1.body, put2.body) {
		t.Errorf("Parsing of Body (put)  does not work")
	}
}

//this test function sends a dhtPUT request and then a dhtGET request to receive the fitting dhtSUCCESS message
//afterwards it sends a second dhtGET request for a (probably) not exiting/stored key to retreive a dhtFailure message
func TestAPICommunication(t *testing.T) {
	/* The following three lines of code are needed, if you want to run this Test function on its own
	go main()
	testMap = make(map[id][]byte)
	time.Sleep(1 * time.Second)

	*/

	waitingTime := time.Duration(ran.Intn(1000))

	testAddr := Conf.apiIP + ":" + strconv.Itoa(int(Conf.apiPort))

	fmt.Println("[TEST] Start of API Test-Programm...")
	tcpAddr, err := net.ResolveTCPAddr("tcp", testAddr)
	if err != nil {
		fmt.Println("Resolving of TCP Address failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Resolved TCP Address: ", tcpAddr)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("[TEST] Dial failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Connection established...")

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)
	thisNode.thisPeer.id = i

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	putBdy := putBody{
		ttl:         20,
		reserved:    2,
		replication: 3,
		key:         i,
		value:       value,
	}

	putMsg := makeApiMessageOutOfBody(&putBdy, dhtPUT)

	_, err = conn.Write(putMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Wrote a dhtPUT message to dht Instance...", waitingTime, ": ", putMsg.body.(*putBody).value)
	fmt.Println()
	time.Sleep(time.Duration(waitingTime * time.Millisecond))
	time.Sleep(800 * time.Millisecond)
	getBdy := getBody{key: i}
	getMsg := makeApiMessageOutOfBody(&getBdy, dhtGET)

	_, err = conn.Write(getMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Wrote a dhtGet request to dht Instance...", waitingTime)
	//time.Sleep(1 * time.Second)
	fmt.Println()

	reply := make([]byte, maxMessageLength)
	msgSize, err := conn.Read(reply)
	if err != nil {
		fmt.Println("[TEST] Write to server failed:", err.Error())
		//os.Exit(1)
		return
	}
	answerMsg := makeApiMessageOutOfBytes(reply[:msgSize])
	fmt.Println()
	//fmt.Println("[TEST] received this answer: ", answerMsg.toString(), waitingTime)
	fmt.Println("[TEST] received an answer: ", waitingTime)

	if answerMsg.header.messageType != dhtSUCCESS {
		t.Errorf("[FAILURE] We did not receive a dhtSUCCESS answer")
	}
	if !reflect.DeepEqual(answerMsg.body.(*successBody).value, putMsg.body.(*putBody).value) {
		t.Errorf("[FAILURE] Returned answer value is not what we asked to be stored")

	} else {
		fmt.Println("[TEST] SUCCESS - we received the right value back", waitingTime, ": ", answerMsg.body.(*successBody).value)
		counter++
	}

	time.Sleep(waitingTime * time.Millisecond)
	fmt.Println()
	getBdy.key[1] = 0
	getBdy.key[4] = 0
	getBdy.key[20] = 0
	getMsg2 := makeApiMessageOutOfBody(&getBdy, dhtGET)
	_, err = conn.Write(getMsg2.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		//os.Exit(1)
		return
	}
	fmt.Println("[TEST] Wrote a dhtGet request (for non-existing key to dht Instance)...", waitingTime)
	//time.Sleep(1 * time.Second)
	fmt.Println()

	reply2 := make([]byte, maxMessageLength)
	msgSize2, err := conn.Read(reply2)
	if err != nil {
		fmt.Println("[TEST] Write to server failed:", err.Error())
		//os.Exit(1)
		return
	}
	answerMsg2 := makeApiMessageOutOfBytes(reply2[:msgSize2])
	fmt.Println()
	//fmt.Println("[TEST] received this answer: ", answerMsg2.toString())
	fmt.Println("[TEST] received an answer (2nd): ", waitingTime)

	if answerMsg2.header.messageType != dhtFAILURE {
		t.Errorf("[FAILURE] We did not receive a dhtSUCCESS answer. (there is a small probability that the sent out key equals the randomly generated key from the first run)")
	} else {
		counter++
	}

}

func TestAPICommunicationConcurrency(t *testing.T) {
	go main()
	testMap = make(map[id][]byte)
	time.Sleep(1 * time.Second)
	numberOfConcurrentTests := 1000
	for i := 0; i < numberOfConcurrentTests; i++ {
		go TestAPICommunication(t)
	}
	time.Sleep(25 * time.Second)
	fmt.Println(counter, "out of ", numberOfConcurrentTests*2, " tests did work")
}

var counter int
