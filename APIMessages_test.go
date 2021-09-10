package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

/*
makeApiMessageOutOfBody is only needed for testing purposes.
It generates instances of ApiMessages provided with an apiBody e.g. of dhtGet, dhtPut, dhtSuccess, dhtFailure
For production the makeApiMessageOutOfAnswer() function suffices
*/
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

	randomBytesForID := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForID); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], randomBytesForID)
	thisNode.thisPeer.id = i

	randomBytesForKey := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForKey); err != nil {
		panic(err.Error())
	}
	var key id
	copy(key[:], randomBytesForKey)

	//first we create a dhtGet Message
	getBdy := getBody{key: key}
	get1 := makeApiMessageOutOfBody(&getBdy, dhtGET)
	fmt.Println("Get_1: ", get1.toString())
	if get1.body == nil {
		t.Errorf("Body of Get message is  nil")
	}

	fmt.Println(get1.data)

	//second we create another dhtGet Message out of the byte representation of the first dhtGet message
	get2 := makeApiMessageOutOfBytes(get1.data)
	fmt.Println("get2: ", get2.toString())

	//third we compare both messages to see if they are the identical
	if get1.header.size != get2.header.size {
		t.Errorf("[FAILURE] Parsing of Header size does not work")
	}
	if get1.header.messageType != get2.header.messageType {
		t.Errorf("[FAILURE] Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(get1.body, get2.body) {
		t.Errorf("[FAILURE] Parsing of Body (get)  does not work")
	}

}

func TestFailureCodingAndDecoding(t *testing.T) {

	randomBytesForID := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForID); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], randomBytesForID)
	thisNode.thisPeer.id = i

	randomBytesForKey := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForKey); err != nil {
		panic(err.Error())
	}
	var key id
	copy(key[:], randomBytesForKey)

	failureBdy := failureBody{key: key}

	//first we create a dhtFailure  Message
	failure1 := makeApiMessageOutOfBody(&failureBdy, dhtFAILURE)
	fmt.Println("Failure1: ", failure1.toString())
	if failure1.body == nil {
		t.Errorf("[FAILURE] Body of Failure message is  nil")
	}

	fmt.Println(failure1.data)
	//second we create another dhtFailure Message out of the byte representation of the first dhtFailure message
	failure2 := makeApiMessageOutOfBytes(failure1.data)
	fmt.Println("failure2: ", failure2.toString())

	//third we compare both messages to see if they are the identical
	if failure1.header.size != failure2.header.size {
		t.Errorf("[FAILURE] Parsing of Header size does not work")
	}
	if failure1.header.messageType != failure2.header.messageType {
		t.Errorf("[FAILURE] Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(failure1.body, failure2.body) {
		t.Errorf("[FAILURE] Parsing of Body (failure)  does not work")
	}

}

func TestSuccessCodingAndDecoding(t *testing.T) {

	randomBytesForID := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForID); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], randomBytesForID)
	thisNode.thisPeer.id = i

	randomBytesForKey := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForKey); err != nil {
		panic(err.Error())
	}
	var key id
	copy(key[:], randomBytesForKey)

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	successBdy := successBody{
		key:   key,
		value: value,
	}

	//first we create a dhtSuccess Message
	success1 := makeApiMessageOutOfBody(&successBdy, dhtSUCCESS)
	fmt.Println("Success1: ", success1.toString())
	if success1.body == nil {
		t.Errorf("Body of Success message is  nil")
	}

	fmt.Println(success1.data)
	//second we create another dhtSuccess Message out of the byte representation of the first dhtSuccess message
	success2 := makeApiMessageOutOfBytes(success1.data)
	fmt.Println("success2: ", success2.toString())

	//third we compare both messages to see if they are the identical
	if success1.header.size != success2.header.size {
		t.Errorf("[FAILURE] Parsing of Header size does not work")
	}
	if success1.header.messageType != success2.header.messageType {
		t.Errorf("[FAILURE] Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(success1.body, success2.body) {
		t.Errorf("[FAILURE] Parsing of Body (success)  does not work")
	}

}

func TestPutCodingAndDecoding(t *testing.T) {

	randomBytesForID := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForID); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], randomBytesForID)
	thisNode.thisPeer.id = i

	randomBytesForKey := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(randomBytesForKey); err != nil {
		panic(err.Error())
	}
	var key id
	copy(key[:], randomBytesForKey)

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	putBdy := putBody{
		ttl:         20,
		reserved:    2,
		replication: 3,
		key:         key,
		value:       value,
	}
	//first we create a dhtPut Message

	put1 := makeApiMessageOutOfBody(&putBdy, dhtPUT)
	fmt.Println("put1: ", put1.toString())
	if put1.body == nil {
		t.Errorf("[FAILURE] Body of Put message is  nil")
	}

	fmt.Println(put1.data)
	//second we create another dhtPut Message out of the byte representation of the first dhtPut message

	put2 := makeApiMessageOutOfBytes(put1.data)
	fmt.Println("put2: ", put2.toString())

	//third we compare both messages to see if they are the identical
	if put1.header.size != put2.header.size {
		t.Errorf("[FAILURE] Parsing of Header size does not work")
	}
	if put1.header.messageType != put2.header.messageType {
		t.Errorf("[FAILURE] Parsing of Header messageType does not work")
	}
	if !reflect.DeepEqual(put1.body, put2.body) {
		t.Errorf("[FAILURE] Parsing of Body (put)  does not work")
	}
}
