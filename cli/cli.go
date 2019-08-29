package cli

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

type Cli struct {
	// Fabric network information
	ConfigPath string

	CCID      string   // chaincode ID, eq name
	CCPath    string   // chaincode source path, 是GOPATH下的某个目录
	CCGoPath  string   // GOPATH used for chaincode
	CCVersion int      // chaincode version
	CCPolicy  string   // endorser policy
	CCpeers   []string // peers used by chaincode

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
		CCID:       "gocc7",
		CCPath:     "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go/", // 相对路径是从GOPAHT/src开始的
		CCGoPath:   os.Getenv("GOPATH"),
		CCVersion:  0,
		CCpeers:    []string{"peer0.org1.example.com"}, // TODO fill peers url, get from config
		// CCPolicy:   "ANY",
		CCPolicy:  "OR('Org1MSP.member','Org2MSP.member')",
		OrgName:   "Org1",
		OrgAdmin:  "Admin",
		OrgUser:   "User1",
		ChannelID: "mychannel",
	}

	// create sdk
	sdk, err := fabsdk.New(config.FromFile(c.ConfigPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	c.Sdk = sdk
	log.Println("Initialized fabric sdk")

	// create rc
	rcp := sdk.Context(fabsdk.WithUser(c.OrgAdmin), fabsdk.WithOrg(c.OrgName))
	rc, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed to create resource client: %s", err)
	}
	c.RC = rc
	log.Println("Initialized resource client")

	// create cc
	ccp := sdk.ChannelContext(c.ChannelID, fabsdk.WithUser(c.OrgUser))
	cc, err := channel.New(ccp)
	if err != nil {
		log.Panicf("failed to create channel client: %s", err)
	}
	c.CC = cc
	log.Println("Initialized channel client")

	return c
}

func (c *Cli) InstallCC(v int) error {
	reqPeers := resmgmt.WithTargetEndpoints(c.CCpeers...)

	// TODO 如果已安装则返回
	// resp, err := c.RC.QueryInstalledChaincodes(reqPeers)
	// if err != nil {
	// }
	// // resp.

	// pack the chaincode
	ccPkg, err := gopackager.NewCCPackage(c.CCPath, c.CCGoPath)
	if err != nil {
		return errors.WithMessage(err, "pack chaincode error")
	}

	// new request of installing chaincode
	req := resmgmt.InstallCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: strconv.Itoa(v),
		Package: ccPkg,
		// send request
	}

	resps, err := c.RC.InstallCC(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "installCC error")
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
		// TODO 增加已安装查询后，恢复返回错误
		// return errors.WithMessage(errs[0], "installCC first error")
	}
	return nil
}

func (c *Cli) InstantiateCC() error {
	reqPeers := resmgmt.WithTargetEndpoints(c.CCpeers...)

	// endorser policy
	ccPolicy, err := c.genPolicy(c.CCPolicy)
	if err != nil {
		return errors.WithMessage(err, "gen policy from string error")
	}
	log.Printf("Instantiate endorser policy: %v", ccPolicy.GetRule().String())

	// new request
	// Attention: args should include `init` for Request not
	// have a method term to call init
	args := packArgs([]string{"init", "a", "100", "b", "200"})
	req := resmgmt.InstantiateCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: strconv.Itoa(c.CCVersion),
		Args:    args,
		Policy:  ccPolicy,
	}

	// send request and handle response
	resp, err := c.RC.InstantiateCC(c.ChannelID, req, reqPeers)
	if err != nil {
		// TODO 已经实例化，增加Invoke前的查询操作后删除
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return errors.WithMessage(err, "instantiate chaincode error")
	}

	log.Printf("Instantitate chaincode tx: %s", resp.TransactionID)
	return nil
}

func (c *Cli) genPolicy(p string) (*common.SignaturePolicyEnvelope, error) {
	// TODO any bug
	if p == "ANY" {
		return cauthdsl.SignedByAnyMember([]string{c.OrgName}), nil
	}
	return cauthdsl.FromString(c.CCPolicy)
}

func (c *Cli) InvokeCC(peers []string) error {
	reqPeers := channel.WithTargetEndpoints(peers...)

	// new channel request for invoke
	args := packArgs([]string{"a", "b", "10"})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "invoke",
		Args:        args,
	}

	// send request and handle response
	// TODO 不设置peer，是否会自动选择peer进行背书, NO
	resp, err := c.CC.Execute(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "invoke chaincode error")
	}

	log.Printf("invoke chaincode tx: %s", resp.TransactionID)
	return nil
}

func (c *Cli) QueryCC(keys ...string) error {
	reqPeers := channel.WithTargetEndpoints(c.CCpeers...)

	// new channel request for query
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "query",
		Args:        packArgs(keys),
	}

	// send request and handle response
	resp, err := c.CC.Query(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "query chaincode error")
	}

	log.Printf("query chaincode tx: %s", resp.TransactionID)
	log.Printf("result: %v", string(resp.Payload))
	return nil
}

func (c *Cli) UpgradeCC(v int) error {
	reqPeers := resmgmt.WithTargetEndpoints(c.CCpeers...)

	// endorser policy
	org1AndOrg2 := "AND('Org1MSP.member','Org2MSP.member')"
	ccPolicy, err := c.genPolicy(org1AndOrg2)
	if err != nil {
		return errors.WithMessage(err, "gen policy from string error")
	}
	log.Printf("Upgrade endorser policy: %v", ccPolicy.String())

	// new request
	// Attention: args should include `init` for Request not
	// have a method term to call init
	args := packArgs([]string{"init", "a", "100", "b", "200"})
	req := resmgmt.UpgradeCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: strconv.Itoa(v),
		Args:    args,
		Policy:  ccPolicy,
	}

	// send request and handle response
	resp, err := c.RC.UpgradeCC(c.ChannelID, req, reqPeers)
	if err != nil {
		// TODO 已经实例化，增加Invoke前的查询操作后删除
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return errors.WithMessage(err, "instantiate chaincode error")
	}

	log.Printf("Instantitate chaincode tx: %s", resp.TransactionID)
	return nil
}

func packArgs(paras []string) [][]byte {
	var args [][]byte
	for _, k := range paras {
		args = append(args, []byte(k))
	}
	return args
}
