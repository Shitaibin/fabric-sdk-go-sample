package main

import (
	"log"

	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const (
	org1CfgPath = "./config/org1sdk-config.yaml"
	org2CfgPath = "./config/org2sdk-config.yaml"
)

var (
	peers1 = []string{"peer0.org1.example.com"}
	peers2 = []string{"peer0.org2.example.com"}
)

func main() {
	org1 := cli.New(org1CfgPath, "Org1", "Admin", "User1")
	org2 := cli.New(org2CfgPath, "Org2", "Admin", "User1")

	// Install, instantiate, invoke, query
	Phase1(org1, org2)
	// Install, upgrade, invoke, query
	Phase2(org1, org2)
}

func Phase1(cli1, cli2 *cli.Cli) {
	log.Println("=================== Phase 1 begin ===================")
	defer log.Println("=================== Phase 1 end ===================\n\n")

	if err := cli1.InstallCC("v1", peers1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org1's peer\n\n")

	if err := cli2.InstallCC("v1", peers2); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org2's peer\n\n")

	// InstantiateCC chaincode only need once for each channel
	if err := cli1.InstantiateCC("v1", peers1); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated\n\n")

	if err := cli1.InvokeCC(peers1); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success\n\n")

	if err := cli1.QueryCC("peer0.org1.example.com", "a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
	log.Println("Query chaincode success on peer0.org1")
}

func Phase2(cli1, cli2 *cli.Cli) {
	log.Println("=================== Phase 2 begin ===================")
	defer log.Println("=================== Phase 2 end ===================\n\n")

	v := "v2"

	// Install new version chaincode
	if err := cli1.InstallCC(v, peers1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org1's peer\n\n")

	if err := cli2.InstallCC(v, peers2); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org2's peer\n\n")

	// Upgrade chaincode only need once for each channel
	if err := cli1.UpgradeCC(v, peers1); err != nil {
		log.Panicf("Upgrade chaincode error: %v", err)
	}
	log.Println("Upgrade chaincode success for channel")

	if err := cli1.InvokeCC([]string{"peer0.org1.example.com", "peer0.org2.example.com"}); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success")

	if err := cli1.QueryCC("peer0.org2.example.com", "a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
	log.Println("Query chaincode success on peer0.org2")
}
