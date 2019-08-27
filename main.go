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
		log.Panic(err)
	}

	// Instantiate
	if err := c.InstantiateCC(); err != nil {
		log.Panic(err)
	}

	// Query

	// Invoke
	if err := c.InvokeCC(); err != nil {
		log.Panic(err)
	}
}
