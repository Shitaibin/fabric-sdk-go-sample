package main

import (
	"log"

	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const cfgPath = "./config/config.yaml"

func main() {
	c := cli.New(cfgPath)
	// Install, instantiate, invoke, query
	Phase1(c)
	// Install, upgrade, invoke, query
	Phase2(c)
}

func Phase1(c *cli.Cli) {
	log.Println("=================== Phase 1 begin ===================")
	defer log.Println("=================== Phase 1 end ===================\n\n")

	log.Println("Installing chaincode v1 to peer0.org1")
	if err := c.InstallCC(c.CCVersion); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed\n\n")

	log.Println("Instantiating chaincode v1")
	if err := c.InstantiateCC(); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated\n\n")

	log.Println("Invoking chaincode on peer0.org1")
	if err := c.InvokeCC(c.CCpeers); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success\n\n")

	log.Println("Query chaincode on peer0.org1")
	if err := c.QueryCC("a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
	log.Println("Query chaincode success")
}

func Phase2(c *cli.Cli) {
	log.Println("=================== Phase 2 begin ===================")
	defer log.Println("=================== Phase 2 end ===================\n\n")

	log.Println("Installing chaincode v2 to peer0.org1, peer0.org2")
	if err := c.InstallCC(c.CCVersion + 1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed")

	log.Println("Upgrade chaincode v2")
	if err := c.UpgradeCC(c.CCVersion + 1); err != nil {
		log.Panicf("Upgrade chaincode error: %v", err)
	}
	log.Println("Upgrade chaincode success")

	// // Invoke
	// if err := c.InvokeCC([]string{"peer0.org1.example.com", "peer0.org2.example.com"}); err != nil {
	// 	log.Panicf("Invoke chaincode error: %v", err)
	// }
	// log.Println("Invoke chaincode success")
}
