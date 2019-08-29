package main

import (
	"log"

	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const cfgPath = "./config/config.yaml"

func main() {
	c := cli.New(cfgPath)
	Phase1(c)
	Phase2(c)
}

func Phase1(c *cli.Cli) {
	// Install
	if err := c.InstallCC(c.CCVersion); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed")

	// Instantiate
	if err := c.InstantiateCC(); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated")

	// Invoke
	if err := c.InvokeCC(c.CCpeers); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success")

	// 	Query `a`
	if err := c.QueryCC("a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
}

func Phase2(c *cli.Cli) {
	// 	Install new version, then upgrade
	if err := c.InstallCC(c.CCVersion + 1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed")

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
