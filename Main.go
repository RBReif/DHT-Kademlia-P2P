package main

import (
	//"DHT-16/api"
	//"DHT-16/p2p"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
	"sync"
)

var Conf configuraton

func parseConfig() configuraton {
	var pathToConfigFile string
	flag.StringVar(&pathToConfigFile, "c", "config/config1.ini", "Specify the path to the config file")
	config, err := ini.Load(pathToConfigFile)
	if err != nil {
		fmt.Println("Could not parse specified config file.")
		os.Exit(1)
	}

	var lock = &sync.Mutex{}
	lock.Lock()
	defer lock.Unlock()
	tmpMaxTtl, err := config.Section("dht").Key("maxTTL").Int()
	if err != nil {
		fmt.Println("Wrong configuration: maxTTL is not an Integer")
		os.Exit(1)
	}
	tmpMinRep, err := config.Section("dht").Key("minReplication").Int()
	if err != nil {
		fmt.Println("Wrong configuration: minReplication is not an Integer")
		os.Exit(1)
	}
	tmpMaxRep, err := config.Section("dht").Key("maxReplication").Int()
	if err != nil {
		fmt.Println("Wrong configuration: maxReplication is not an Integer")
		os.Exit(1)
	}
	k, err := config.Section("dht").Key("k").Int()
	if err != nil {
		fmt.Println("Wrong configuration: maxReplication is not an Integer")
		os.Exit(1)
	}
	a, err := config.Section("dht").Key("a").Int()
	if err != nil {
		fmt.Println("Wrong configuration: maxReplication is not an Integer")
		os.Exit(1)
	}

	apiPort, err := strconv.Atoi(strings.Split(config.Section("dht").Key("api_address").String(), ":")[1])
	if err != nil {
		fmt.Println("Wrong configuration: the port of the apiAdress is not an Integer")
		os.Exit(1)
	}
	p2pPort, err := strconv.Atoi(strings.Split(config.Section("dht").Key("p2p_address").String(), ":")[1])
	if err != nil {
		fmt.Println("Wrong configuration: the port of the p2pAdress is not an Integer")
		os.Exit(1)
	}

	conf := configuraton{
		HostKeyFile: config.Section("").Key("hostkey").String(),
		apiIP:       strings.Split(config.Section("dht").Key("api_address").String(), ":")[0],
		apiPort:     uint16(apiPort),
		p2pIP:       strings.Split(config.Section("dht").Key("p2p_address").String(), ":")[0],
		p2pPort:     uint16(p2pPort),

		//minTTL:           time.Duration(),
		maxTTL:         tmpMaxTtl,
		minReplication: tmpMinRep,
		maxReplication: tmpMaxRep,
		preConfPeer1:   config.Section("dht").Key("preConfPeer1").String(),
		preConfPeer2:   config.Section("dht").Key("preConfPeer2").String(),
		preConfPeer3:   config.Section("dht").Key("preConfPeer3").String(),
		//apiAddressRPS:  config.Section("rps").Key("api_address").String(),
		k: k,
		a: a,
	}
	if !conf.checkConfig() {
		fmt.Println("Wrong configuration: an address is wrongly configured")
		os.Exit(1)
	}
	fmt.Println("Read and Parsed the following Configuration file: ", conf.toString())
	return conf
}

func main() {
	fmt.Println("Program started...")
	Conf = parseConfig()
	go startAPIDispatcher()
	initialize()

}

type configuraton struct {
	//general
	HostKeyFile string
	//dht
	apiIP          string
	apiPort        uint16
	p2pIP          string
	p2pPort        uint16
	maxTTL         int
	minReplication int
	maxReplication int
	preConfPeer1   string
	preConfPeer2   string
	preConfPeer3   string
	//kademlia specific
	k int
	a int
	//rps
	//apiAddressRPS string
}

func (c *configuraton) toString() string {
	str := "Configuration file: "
	str = str + "   HostKeyFile: " + c.HostKeyFile + "\n"
	str = str + "   apiIP: " + c.apiIP + "\n"
	str = str + "   apiPort: " + strconv.Itoa(int(c.apiPort)) + "\n"
	str = str + "   p2pIP: " + c.p2pIP + "\n"
	str = str + "   p2pPort: " + strconv.Itoa(int(c.p2pPort)) + "\n"
	str = str + "   maxTTL: " + strconv.Itoa(c.maxTTL) + "\n"
	str = str + "   minReplication: " + strconv.Itoa(c.minReplication) + "\n"
	str = str + "   maxReplication: " + strconv.Itoa(c.maxReplication) + "\n"
	str = str + "   preConfPeer1: " + c.preConfPeer1 + "\n"
	str = str + "   preConfPeer2: " + c.preConfPeer2 + "\n"
	str = str + "   preConfPeer3: " + c.preConfPeer3 + "\n"
	//str = str + "   apiAddressRPS: " + c.apiAddressRPS + "\n"
	return str
}
func (c *configuraton) checkConfig() bool {
	everythingAlright := true
	if strings.Contains(c.apiIP, ":") || strings.Contains(c.p2pIP, ":") {
		everythingAlright = false
	}
	return everythingAlright
}
