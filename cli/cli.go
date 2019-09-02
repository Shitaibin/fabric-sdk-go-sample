package cli

import (
	"log"
	"net/http"
	"os"
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
	OrgName    string
	OrgAdmin   string
	OrgUser    string

	// sdk clients
	sdk *fabsdk.FabricSDK
	rc  *resmgmt.Client
	cc  *channel.Client

	// Same for each peer
	ChannelID string
	CCID      string // chaincode ID, eq name
	CCPath    string // chaincode source path, 是GOPATH下的某个目录
	CCGoPath  string // GOPATH used for chaincode
}

func New(cfg, org, admin, user string) *Cli {
	c := &Cli{
		ConfigPath: cfg,
		OrgName:    org,
		OrgAdmin:   admin,
		OrgUser:    user,

		CCID:      "gocc9",
		CCPath:    "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go/", // 相对路径是从GOPAHT/src开始的
		CCGoPath:  os.Getenv("GOPATH"),
		ChannelID: "mychannel",
	}

	// create sdk
	sdk, err := fabsdk.New(config.FromFile(c.ConfigPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	c.sdk = sdk
	log.Println("Initialized fabric sdk")

	c.rc, c.cc = NewSdkClient(sdk, c.ChannelID, c.OrgName, c.OrgAdmin, c.OrgUser)
	return c
}

// NewSdkClient create resource client and channel client
func NewSdkClient(sdk *fabsdk.FabricSDK, channelID, orgName, orgAdmin, OrgUser string) (rc *resmgmt.Client, cc *channel.Client) {
	var err error

	// create rc
	rcp := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	rc, err = resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed to create resource client: %s", err)
	}
	log.Println("Initialized resource client")

	// create cc
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser(OrgUser))
	cc, err = channel.New(ccp)
	if err != nil {
		log.Panicf("failed to create channel client: %s", err)
	}
	log.Println("Initialized channel client")

	return rc, cc
}

// TODO 每个节点需要单独安装，
func (c *Cli) InstallCC(v string, peers []string) error {
	// TODO 如果peer已安装则跳过

	// pack the chaincode
	ccPkg, err := gopackager.NewCCPackage(c.CCPath, c.CCGoPath)
	if err != nil {
		return errors.WithMessage(err, "pack chaincode error")
	}

	// new request of installing chaincode
	req := resmgmt.InstallCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: v,
		Package: ccPkg,
	}

	reqPeers := resmgmt.WithTargetEndpoints(peers...)
	resps, err := c.rc.InstallCC(req, reqPeers)
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

func (c *Cli) InstantiateCC(v string, peers []string) error {
	// endorser policy
	org1OrOrg2 := "OR('Org1MSP.member','Org2MSP.member')"
	ccPolicy, err := c.genPolicy(org1OrOrg2)
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
		Version: v,
		Args:    args,
		Policy:  ccPolicy,
	}

	// send request and handle response
	reqPeers := resmgmt.WithTargetEndpoints(peers...)
	resp, err := c.rc.InstantiateCC(c.ChannelID, req, reqPeers)
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
	// TODO bug, this any leads to endorser invalid
	if p == "ANY" {
		return cauthdsl.SignedByAnyMember([]string{c.OrgName}), nil
	}
	return cauthdsl.FromString(p)
}

func (c *Cli) InvokeCC(peers []string) error {
	// new channel request for invoke
	args := packArgs([]string{"a", "b", "10"})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "invoke",
		Args:        args,
	}

	// send request and handle response
	// peers is needed
	reqPeers := channel.WithTargetEndpoints(peers...)
	resp, err := c.cc.Execute(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "invoke chaincode error")
	}

	log.Printf("invoke chaincode tx: %s", resp.TransactionID)
	return nil
}

func (c *Cli) QueryCC(peer, keys string) error {
	// new channel request for query
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "query",
		Args:        packArgs([]string{keys}),
	}

	// send request and handle response
	reqPeers := channel.WithTargetEndpoints(peer)
	resp, err := c.cc.Query(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "query chaincode error")
	}

	log.Printf("query chaincode tx: %s", resp.TransactionID)
	log.Printf("result: %v", string(resp.Payload))
	return nil
}

func (c *Cli) UpgradeCC(v string, peers []string) error {
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
		Version: v,
		Args:    args,
		Policy:  ccPolicy,
	}

	// send request and handle response
	reqPeers := resmgmt.WithTargetEndpoints(peers...)
	resp, err := c.rc.UpgradeCC(c.ChannelID, req, reqPeers)
	if err != nil {
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
