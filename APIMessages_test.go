package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
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
