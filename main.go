package main

import (
	"log"

	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const cfgPath = "./config/config.yaml"

func main() {
	c := cli.New(cfgPath)

	// Install
	if err := c.InstallCC(); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed")

	// Instantiate
	if err := c.InstantiateCC(); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated")

	// Invoke
	if err := c.InvokeCC(); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success")

	// 	Query `a`
	if err := c.QueryCC("a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
}
