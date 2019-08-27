package cli

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type Cli struct {
	// Fabric network information
	ConfigPath string

	CCID      string // chaincode ID, eq name
	CCPath    string // chaincode source path
	CCGoPath  string // GOPATH used for chaincode
	CCVersion int    // chaincode version

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

func (c *Cli) InstallCC() error {
	peers := []string{} // TODO fill peers url, get from config
	reqPeers := resmgmt.WithTargetEndpoints(peers...)

	// TODO 如果已安装则返回
	// resp, err := c.RC.QueryInstalledChaincodes(reqPeers)
	// if err != nil {
	// 	return fmt.Errorf("failed to query installed cc: %v", err)
	// }
	// // resp.

	// pack the chaincode
	ccPkg, err := gopackager.NewCCPackage(c.CCPath, c.CCGoPath)
	if err != nil {
		return err
	}

	// new request of insalling chaincode
	req := resmgmt.InstallCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: strconv.Itoa(c.CCVersion),
		Package: ccPkg,
		// send request
	}

	resps, err := c.RC.InstallCC(req, reqPeers)
	if err != nil {
		return fmt.Errorf("InstallCC returned error: %v", err)
	}

	// check other errors
	var errs []error
	for _, resp := range resps {
		if resp.Status != http.StatusOK {
			errs = append(errs, errors.New(resp.Info))
		}
	}

	if len(errs) > 0 {
		log.Printf("InstallCC errors: %v", errs)
		return errs[0]
	}
	return nil
}
