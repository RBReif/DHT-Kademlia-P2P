package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"strconv"
	"sync"
)

var Conf configuraton

//This function parses the configuration file that was provided with the -c flag
func parseConfig() configuraton {
	var pathToConfigFile string

	flag.StringVar(&pathToConfigFile, "c", "config/mainConfig.ini", "Specify the path to the config file")
	flag.Parse()

	config, err := ini.Load(pathToConfigFile)
	if err != nil {
		log.Fatal("[FAILURE] Could not parse specified config file.")
	}

	var lock = &sync.Mutex{}
	lock.Lock()
	defer lock.Unlock()
	tmpMaxTtl, err := config.Section("dht").Key("maxTTL").Int()
	if err != nil {
		log.Fatal("[FAILURE] Wrong configuration: maxTTL is not an Integer")
	}

	k, err := config.Section("dht").Key("k").Int()
	if err != nil {
		log.Fatal("[FAILURE] Wrong configuration: k is not an Integer")
	}
	a, err := config.Section("dht").Key("a").Int()
	if err != nil {
		log.Fatal("[FAILURE] Wrong configuration: a is not an Integer")
	}

	apiAddr := extractPeerAddressFromString(config.Section("dht").Key("api_address").String())
	p2pAddr := extractPeerAddressFromString(config.Section("dht").Key("p2p_address").String())

	conf := configuraton{
		HostKeyFile:  config.Section("").Key("hostkey").String(),
		apiIP:        apiAddr.ip,
		apiPort:      apiAddr.port,
		p2pIP:        p2pAddr.ip,
		p2pPort:      p2pAddr.port,
		maxTTL:       tmpMaxTtl,
		preConfPeer1: config.Section("dht").Key("preConfPeer1").String(),
		preConfPeer2: config.Section("dht").Key("preConfPeer2").String(),
		preConfPeer3: config.Section("dht").Key("preConfPeer3").String(),
		k:            k,
		a:            a,
	}

	log.Info("[SUCCESS] Read and Parsed the following Configuration file: ", conf.toString())
	return conf
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	Conf = parseConfig()
	go startAPIMessageDispatcher(&wg)
	go startP2PMessageDispatcher(&wg)
	initializeP2PCommunication()
	go startTimers()

	log.Info("Program started")
	wg.Wait()
	log.Info("Program stopped")
}

type configuraton struct {
	//general
	HostKeyFile string
	//dht
	apiIP        string
	apiPort      uint16
	p2pIP        string
	p2pPort      uint16
	maxTTL       int
	preConfPeer1 string
	preConfPeer2 string
	preConfPeer3 string
	//kademlia specific
	k int
	a int
}

func (c *configuraton) toString() string {
	str := "Configuration file: "
	str = str + "   HostKeyFile: " + c.HostKeyFile + "\n"
	str = str + "   apiIP: " + c.apiIP + "\n"
	str = str + "   apiPort: " + strconv.Itoa(int(c.apiPort)) + "\n"
	str = str + "   p2pIP: " + c.p2pIP + "\n"
	str = str + "   p2pPort: " + strconv.Itoa(int(c.p2pPort)) + "\n"
	str = str + "   maxTTL: " + strconv.Itoa(c.maxTTL) + "\n"
	str = str + "   preConfPeer1: " + c.preConfPeer1 + "\n"
	str = str + "   preConfPeer2: " + c.preConfPeer2 + "\n"
	str = str + "   preConfPeer3: " + c.preConfPeer3 + "\n"
	return str
}
