package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

//listens for TCP connections for API calls
func startAPIMessageDispatcher(wg *sync.WaitGroup) {
	defer wg.Done()

	//we listen on the specified API address from the configuration file
	l, err := net.Listen("tcp", Conf.apiIP+":"+strconv.Itoa(int(Conf.apiPort)))
	if err != nil {
		custError := "[FAILURE] MAIN: Error while listening for connection at" + Conf.apiIP + ": " + strconv.Itoa(int(Conf.apiPort)) + " - " + err.Error()
		fmt.Println(custError)
		panic(custError)
	}
	defer l.Close()
	fmt.Println("[SUCCESS] MAIN: APIMessageDispatcher Listening on ", Conf.apiIP, ": ", Conf.apiPort)

	for {
		con, err := l.Accept()
		if err != nil {
			custError := "[FAILURE] MAIN: Error while accepting: " + err.Error()
			fmt.Println(custError)
			panic(custError)
		}
		fmt.Println("[SUCCESS] MAIN " + strconv.Itoa(int(Conf.apiPort)) + ": New Connection established")
		con.SetDeadline(time.Now().Add(time.Minute * 20)) //Set Timeout

		//for each newly established connection we concurrently call the handleAPIconnection() function
		go handleAPIconnection(con)
	}
}

//listens on one connection for new messages
func handleAPIconnection(con net.Conn) {
	for {
		//On the connection we read the next message
		receivedMessageRaw := make([]byte, maxMessageLength)
		msgSize, err := con.Read(receivedMessageRaw)
		if err != nil {
			custError := "[pot. FAILURE] MAIN: Error while reading from connection: " + err.Error() + " (This might be because no more data was sent)"
			fmt.Println(custError)
			con.Close()
			return
		}
		if msgSize > maxMessageLength {
			custError := "[FAILURE] MAIN: Too much data was sent to us: " + strconv.Itoa(msgSize)
			fmt.Println(custError)
			con.Close()
			return
		}

		size := binary.BigEndian.Uint16(receivedMessageRaw[:2])
		fmt.Println("DEBUG: Received message has size: ", size)
		if uint16(msgSize) != size {
			custError := "[FAILURE] MAIN " + strconv.Itoa(int(Conf.apiPort)) + ": Message size (" + strconv.Itoa(msgSize) + ") does not match specified 'size': " + strconv.Itoa(int(size)) + " sometimes this happens if two messages are sent to quickly)"
			fmt.Println(custError)
			con.Close()
			return
		}

		//out of the received bytes we create an instance of type apiMessage
		receivedMsg := makeApiMessageOutOfBytes(receivedMessageRaw[:msgSize])
		fmt.Println("API ", Conf.apiPort, " Received message : ", receivedMsg.toString())

		switch receivedMsg.header.messageType {
		case dhtPUT:
			handlePut(receivedMsg.body.(*putBody))

		case dhtGET:
			if receivedMsg.header.size != 36 {
				custError := "[FAILURE] MAIN: Message size (" + strconv.Itoa(msgSize) + ") does not match expected size for a GET message"
				fmt.Println(custError)
				con.Close()
				return
			}

			answer := handleGet(receivedMsg.body.(*getBody))

			//the answerMessage will be of type dhtFailure or dhtSuccess
			answerMessage := makeApiMessageOutOfAnswer(answer)

			//we send the anwer back
			_, err := con.Write(answerMessage.data)
			if err != nil {
				custError := "[FAILURE] MAIN:  Error while writing to connection: " + err.Error()
				fmt.Println(custError)
				panic(custError)
			}
			fmt.Println("[SUCCESS] MAIN: Written answer to connection")

		default:
			custError := "[FAILURE] MAIN: Message was of not specified type: " + strconv.Itoa(int(receivedMsg.header.messageType))
			fmt.Println(custError)
			con.Close()
			return
		}

		con.SetDeadline(time.Now().Add(time.Minute * 20)) //Timeout restarted
	}
}

/*
The handleGet function calls the nodeLookup() function according to the Kademlia protocol. In multiple rounds nodeLookup() contacts
peers it believes to be close to the specified key for which we shall retreive the value. In case the value is retreived, nodeLookup()
stores the key-value pair in the local hashTable of this peer. After performing the nodeLookup() this peer can thus
read the key-value pair (if it was found)
*/
func handleGet(body *getBody) DhtAnswer {
	key := body.key
	thisNode.nodeLookup(key, true)
	var value, valueFound = thisNode.hashTable.read(key)
	if valueFound {
		// the value was found. A DHTsuccess message will be sent back
		return DhtAnswer{
			success: true,
			key:     body.key,
			value:   value,
		}
	} else {
		// the value was not found. A DHTFailure message will be sent back
		return DhtAnswer{
			success: false,
			key:     body.key,
		}

	}
}

/*
The handlePut() function first locates the k closest Nodes in network with the nodeLookup() function.
Then it sends KDM_STORE messages with the key-value pair to the k closest nodes to the specified key
*/
func handlePut(body *putBody) {
	// DEBUG: fmt.Println("handlePut has received :", body.toString())

	// store on network
	store(body.key, body.value, body.ttl)

	/*
		as a chaching mechanism we additionally store the key-value pair locally
		in case it is requested briefly again.
	*/
	thisNode.hashTable.write(body.key, body.value, time.Now().Add(time.Duration(body.ttl)*time.Second), time.Now().Add(time.Duration(REPUBLISH_TIME)*time.Second))
}