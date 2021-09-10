package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"testing"
	"time"
)

//please first run the "runPeersConcurrent.sh" bash Skript to start up as many peers concurrently as you want

//this test function sends a dhtPUT request and then a dhtGET request to receive the fitting dhtSUCCESS message
//afterwards it sends a second dhtGET request for a (probably) not exiting/stored key to retreive a dhtFailure message

func helpTestComplete(t *testing.T, wg *sync.WaitGroup) {

	time.Sleep(30 * time.Second)

	waitingTime := time.Duration(200)

	apiAddr := "127.0.0.1:3009"

	fmt.Println("[TEST] Start of API Test-Programm...")
	/*
		tcpAddr, err := net.ResolveTCPAddr("tcp", apiAddr)
		if err != nil {
			fmt.Println("Resolving of TCP Address failed:", err.Error())
			os.Exit(1)
		}
		fmt.Println("[TEST] Resolved TCP Address: ", tcpAddr)
	*/
	conn, err := net.Dial("tcp", apiAddr)
	if err != nil {
		fmt.Println("[TEST] Dial failed:", err.Error())
		t.Errorf("we could not connect to API")
		os.Exit(1)
	}
	fmt.Println("[TEST] Connection established...")

	idx := make([]byte, SIZE_OF_ID)
	if _, err := rand.Read(idx); err != nil {
		panic(err.Error())
	}
	var i id
	copy(i[:], idx)

	value := make([]byte, 10)
	if _, err := rand.Read(value); err != nil {
		panic(err.Error())
	}
	putBdy := putBody{
		ttl:         2000,
		reserved:    2,
		replication: 3,
		key:         i,
		value:       value,
	}

	putMsg := makeApiMessageOutOfBody(&putBdy, dhtPUT)
	fmt.Println("[TEST] we created this message: ", putMsg.toString())
	_, err = conn.Write(putMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		t.Errorf("we could not write a PUT to API ")

		os.Exit(1)
	}
	fmt.Println("[TEST] Wrote a dhtPUT message to dht Instance...", waitingTime, ": ", putMsg.body.toString())
	fmt.Println()
	time.Sleep(time.Duration(waitingTime * time.Millisecond))
	fmt.Println("WAITING.... ")
	time.Sleep(30 * time.Second)
	getBdy := getBody{key: i}
	getMsg := makeApiMessageOutOfBody(&getBdy, dhtGET)

	_, err = conn.Write(getMsg.data)

	if err != nil {
		println("[TEST] Write to dhtInstance failed:", err.Error())
		t.Errorf("we could not write a GET request to API ")

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
	time.Sleep(25 * time.Second)
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
		fmt.Println("Received (as expected) a dhtFAILURE answer")
	}

	wg.Done()
}

var cmd *exec.Cmd

func TestComplete(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go helpTestComplete(t, &wg)

	go help(&wg)
	cmd = exec.Command("/bin/bash", "runPeersConcurrent.sh")
	o, e := cmd.Output()
	fmt.Println("o: ", o)
	fmt.Println("e", e)
	fmt.Println("command is running")
	wg.Wait()
}

func help(wg *sync.WaitGroup) {
	wg.Wait()
	fmt.Println("waiting done")
	err := cmd.Process.Signal(os.Interrupt)
	time.Sleep(1 * time.Second)
	cmd.Process.Kill()
	out, err := exec.Command("pkill", "-f", "DHT-16").Output()
	fmt.Println(out)

	if err != nil {
		fmt.Println(err)
	}
	os.Exit(0)

}
