package main

import (
	"log"

	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const (
	org1CfgPath = "./config/org1sdk-config.yaml"
	org2CfgPath = "./config/org2sdk-config.yaml"
)

func main() {
	org1 := cli.New(org1CfgPath, "Org1", "Admin", "User1")
	// Install, instantiate, invoke, query
	Phase1(org1)
	// Install, upgrade, invoke, query
	Phase2(org1)
}

func Phase1(c *cli.Cli) {
	log.Println("=================== Phase 1 begin ===================")
	defer log.Println("=================== Phase 1 end ===================\n\n")

	peers := []string{"peer0.org1.example.com"}

	log.Println("Installing chaincode v1 to peer0.org1")
	if err := c.InstallCC("v1", peers); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed\n\n")

	log.Println("Instantiating chaincode v1")
	if err := c.InstantiateCC("v1", peers); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated\n\n")

	log.Println("Invoking chaincode on peer0.org1")
	if err := c.InvokeCC(peers); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success\n\n")

	log.Println("Query chaincode on peer0.org1")
	if err := c.QueryCC("peer0.org1.example.com", "a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
	log.Println("Query chaincode success")
}

func Phase2(c *cli.Cli) {
	log.Println("=================== Phase 2 begin ===================")
	defer log.Println("=================== Phase 2 end ===================\n\n")

	peers := []string{"peer0.org1.example.com", "peer0.org2.example.com"}

	log.Println("Installing chaincode v2 to peer0.org1, peer0.org2")
	if err := c.InstallCC("v1", peers); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed")

	// TODO 是否要多节点分开升级
	log.Println("Upgrade chaincode v2")
	if err := c.UpgradeCC("v2", peers); err != nil {
		log.Panicf("Upgrade chaincode error: %v", err)
	}
	log.Println("Upgrade chaincode success")

	// // Invoke
	// if err := c.InvokeCC([]string{"peer0.org1.example.com", "peer0.org2.example.com"}); err != nil {
	// 	log.Panicf("Invoke chaincode error: %v", err)
	// }
	// log.Println("Invoke chaincode success")
}
