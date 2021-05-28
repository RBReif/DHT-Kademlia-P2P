package main

import (
	"DHT-16/api"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
	"sync"
)

var Conf *configuraton

func parseConfig() *configuraton {
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
	conf := &configuraton{
		HostkeyDirectory: config.Section("").Key("hostkey").String(),
		apiAddressDHT:    config.Section("dht").Key("api_address").String(),
		p2pAddressDHT:    config.Section("dht").Key("p2p_address").String(),
		//minTTL:           time.Duration(),
		maxTTL:         tmpMaxTtl,
		minReplication: tmpMinRep,
		maxReplication: tmpMaxRep,
		preConfPeer1:   config.Section("dht").Key("preConfPeer1").String(),
		preConfPeer2:   config.Section("dht").Key("preConfPeer2").String(),
		preConfPeer3:   config.Section("dht").Key("preConfPeer3").String(),
		apiAddressRPS:  config.Section("rps").Key("api_address").String(),
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
	api.Listen(Conf.apiAddressDHT)

}

type configuraton struct {
	//general
	HostkeyDirectory string
	//dht
	apiAddressDHT  string
	p2pAddressDHT  string
	maxTTL         int
	minReplication int
	maxReplication int
	preConfPeer1   string
	preConfPeer2   string
	preConfPeer3   string
	//rps
	apiAddressRPS string
}

func (c *configuraton) toString() string {
	str := "Configuration file: "
	str = str + "   HostkeyDirectory: " + c.HostkeyDirectory + "\n"
	str = str + "   apiAddressDHT: " + c.apiAddressDHT + "\n"
	str = str + "   p2pAddressDHT: " + c.p2pAddressDHT + "\n"
	str = str + "   maxTTL: " + strconv.Itoa(c.maxTTL) + "\n"
	str = str + "   minReplication: " + strconv.Itoa(c.minReplication) + "\n"
	str = str + "   maxReplication: " + strconv.Itoa(c.maxReplication) + "\n"
	str = str + "   preConfPeer1: " + c.preConfPeer1 + "\n"
	str = str + "   preConfPeer2: " + c.preConfPeer2 + "\n"
	str = str + "   preConfPeer3: " + c.preConfPeer3 + "\n"
	str = str + "   apiAddressRPS: " + c.apiAddressRPS + "\n"
	return str
}
func (c *configuraton) checkConfig() bool {
	everythingAlright := true
	if !strings.Contains(c.apiAddressDHT, ":") || !strings.Contains(c.p2pAddressDHT, ":") {
		everythingAlright = false
	}
	return everythingAlright
}
