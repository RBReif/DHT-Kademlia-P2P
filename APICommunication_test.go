package main

import (
	"crypto/rand"
	"fmt"
	ran "math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

/*
This helper function tests primarily our API Communication handling. It connects to the apiAddress. Then it
first sends a dhtPUT request for a newly generated key-value pair to be stored in the system.
Afterwards it sends a dhtGET request for the just used key. It should then receive a dhtSUCCESS message back.
Third, this function also sends a dhtGET request for another key where there should not be any value stored in the system.
We expect to receive a dhtFAILURE message back.

Note: This function is not meant for testing the P2P communication. We only start one peer. This leads to 3 [FAILURE] messages
after initialization as this peer is not able to connect to the bootstraping peers from the configuration file. This is
intended.
*/
func helpTestAPICommunication(t *testing.T) {

	waitingTime := time.Duration(ran.Intn(1000))

	testAddr := Conf.apiIP + ":" + strconv.Itoa(int(Conf.apiPort))

	//First we connect to the apiAddress via tcp
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

	//Second, we generate random bytes as key and value

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

	putMsg := makeApiMessageOutOfBody(&putBdy, dhtPUT)

	_, err = conn.Write(putMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Wrote a dhtPUT message to dht Instance...: ", putMsg.body.(*putBody).value)
	fmt.Println()
	time.Sleep(time.Duration(waitingTime * time.Millisecond))
	time.Sleep(1000 * time.Millisecond)
	getBdy := getBody{key: key}
	getMsg := makeApiMessageOutOfBody(&getBdy, dhtGET)

	_, err = conn.Write(getMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("[TEST] Wrote a dhtGet request to dht Instance...")
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

	if answerMsg.header.messageType != dhtSUCCESS {
		t.Errorf("[FAILURE] We did not receive a dhtSUCCESS answer")
	}
	if !reflect.DeepEqual(answerMsg.body.(*successBody).value, putMsg.body.(*putBody).value) {
		t.Errorf("[FAILURE] Returned answer value is not what we asked to be stored")

	} else {
		fmt.Println("[TEST] SUCCESS - we received the right value back", answerMsg.body.(*successBody).value)
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
	fmt.Println("[TEST] Wrote a dhtGet request (for non-existing key to dht Instance)...")
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
	fmt.Println("[TEST] received an answer (2nd): ")

	if answerMsg2.header.messageType != dhtFAILURE {
		t.Errorf("[FAILURE] We did not receive a dhtSUCCESS answer. (there is a small probability that the sent out key equals the randomly generated key from the first run)")
	} else {
		fmt.Println("[TEST] SUCCESS: received a dhtFailure message (as expected) answer (2nd): ")

		counter++
	}

}

/*
TestAPICommunciationConcurrency runs multiple (e.g. 1000) instances of the helpTestAPICommunication() function
to test the ability to handle hundreds or thousands of concurrent api requests
*/
func TestAPICommunicationConcurrency(t *testing.T) {
	go main()
	time.Sleep(1 * time.Second)
	numberOfConcurrentTests := 1000
	for i := 0; i < numberOfConcurrentTests; i++ {
		go helpTestAPICommunication(t)
	}
	time.Sleep(25 * time.Second) //we now wait a bit to let the multiple tests run
	fmt.Println(counter, "out of ", numberOfConcurrentTests*2, " tests did work")
}

var counter int
