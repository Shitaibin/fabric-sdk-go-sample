package cli

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"log"
	"os"
)

type Cli struct {
	// Fabric network inforation
	ConfigPath string

	CCID      string // Chaincode ID, eq name
	CCPath    string // chaincode source path
	CCGoPath  string
	CCVersion int

	OrgName  string
	OrgAdmin string
	OrgUser  string

	ChannelID string
	// ChannelConfigPath string // used when create a channel

	// SDK
	Sdk *fabsdk.FabricSDK
	RC  *resmgmt.Client
	CC  *channel.Client
}

func New(cfg string) *Cli {
	c := &Cli{
		ConfigPath: cfg,
		CCID:       "chaincode example",
		CCPath:     "github.com/shitaibin/fabric-sdk-go-sample/chaincode",
		CCGoPath:   os.Getenv("GOPATH"),
		CCVersion:  0,
		OrgName:    "Org1",
		OrgAdmin:   "Admin",
		OrgUser:    "User1",
		ChannelID:  "mychannel",
	}

	// create sdk
	sdk, err := fabsdk.New(config.FromFile(c.ConfigPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	c.Sdk = sdk

	// create rc
	rcp := sdk.Context(fabsdk.WithUser(c.OrgAdmin), fabsdk.WithOrg(c.OrgName))
	rc, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed to create resource client: %s", err)
	}
	c.RC = rc

	// create cc
	ccp := sdk.ChannelContext(c.ChannelID, fabsdk.WithUser(c.OrgUser))
	cc, err := channel.New(ccp)
	if err != nil {
		log.Panicf("failed to create channel client: %s", err)
	}
	c.CC = cc

	return c
}
