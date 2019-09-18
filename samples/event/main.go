package main

import (
	"encoding/hex"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/shitaibin/fabric-sdk-go-sample/cli"
)

const (
	org1CfgPath = "../../config/org1sdk-config.yaml"
	org2CfgPath = "../../config/org2sdk-config.yaml"
)

var (
	peer0Org1 = "peer0.org1.example.com"
	peer0Org2 = "peer0.org2.example.com"
)

func main() {
	org1Client := cli.New(org1CfgPath, "Org1", "Admin", "User1")
	org2Client := cli.New(org2CfgPath, "Org2", "Admin", "User1")
	defer org1Client.Close()
	defer org2Client.Close()

	// New event client
	cp := org1Client.SDK.ChannelContext(org1Client.ChannelID,
		fabsdk.WithUser(org1Client.OrgUser))
	ec, err := event.New(cp, event.WithBlockEvents())
	if err != nil {
		log.Printf("Create event client error: %v", err)
	}

	defer ec.Unregister(blockListener(ec))
	defer ec.Unregister(filteredBlockListener(ec))

	// txListener(ec)
	DoChainCode(org1Client, org2Client)
}

func blockListener(ec *event.Client) fab.Registration {
	// Register monitor block event
	beReg, beCh, err := ec.RegisterBlockEvent()
	if err != nil {
		log.Printf("Register block event error: %v", err)
	}

	// Receive block event
	go func() {
		for e := range beCh {
			log.Printf("Receive block event:\nNumber: %v\nHash: %v\nSourceURL"+
				": %v", e.Block.Header.Number, hex.EncodeToString(e.Block.Header.DataHash),
				e.SourceURL)
		}
	}()

	return beReg
}

func filteredBlockListener(ec *event.Client) fab.Registration {
	// Register monitor filtered block event
	fbeReg, fbeCh, err := ec.RegisterFilteredBlockEvent()
	if err != nil {
		log.Printf("Register filtered block event error: %v", err)
	}

	// Receive filtered block event
	go func() {
		for e := range fbeCh {
			log.Printf("Receive filterd block event:\nNumber: %v\nlen("+
				"transactions): %v\nSourceURL: %v",
				e.FilteredBlock.Number, len(e.FilteredBlock.
					FilteredTransactions), e.SourceURL)

			for i, tx := range e.FilteredBlock.FilteredTransactions {
				log.Printf("tx index %d: type: %v, txid: %v, "+
					"validation code: %v", i,
					tx.Type, tx.Txid,
					tx.TxValidationCode)
			}
		}
	}()

	return fbeReg
}

func txListener(ec *event.Client, txIDCh chan string) {
	for id := range txIDCh {
		// Register monitor transaction event
		txReg, txCh, err := ec.RegisterTxStatusEvent(id)
		if err != nil {
			log.Printf("Register transaction event error: %v", err)
		}
		defer ec.Unregister(txReg)

		// Receive transaction event
		go func() {
			for e := range txCh {
				log.Printf("Receive transaction event: txid: %v, "+
					"validation code: %v, block number: %v",
					e.TxID,
					e.TxValidationCode,
					e.BlockNumber)
			}
		}()
	}
}

// Install、Deploy、Invoke、Query、Upgrade
func DoChainCode(org1Client, org2Client *cli.Client) {
	// Install, instantiate, invoke, query
	Phase1(org1Client, org2Client)
	// Install, upgrade, invoke, query
	// Phase2(org1Client, org2Client)
}

func Phase1(cli1, cli2 *cli.Client) {
	log.Println("=================== Phase 1 begin ===================")
	defer log.Println("=================== Phase 1 end ===================")

	if err := cli1.InstallCC("v1", peer0Org1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org1's peers")

	if err := cli2.InstallCC("v1", peer0Org2); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org2's peers")

	// InstantiateCC chaincode only need once for each channel
	if err := cli1.InstantiateCC("v1", peer0Org1); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated")

	if err := cli1.InvokeCC([]string{peer0Org1}); err != nil {
		log.Panicf("Invoke chaincode error: %v", err)
	}
	log.Println("Invoke chaincode success")

	if err := cli1.QueryCC("peer0.org1.example.com", "a"); err != nil {
		log.Panicf("Query chaincode error: %v", err)
	}
	log.Println("Query chaincode success on peer0.org1")
}

// func Phase2(cli1, cli2 *cli.Client) {
// 	log.Println("=================== Phase 2 begin ===================")
// 	defer log.Println("=================== Phase 2 end ===================")
//
// 	v := "v2"
//
// 	// Install new version chaincode
// 	if err := cli1.InstallCC(v, peer0Org1); err != nil {
// 		log.Panicf("Intall chaincode error: %v", err)
// 	}
// 	log.Println("Chaincode has been installed on org1's peers")
//
// 	if err := cli2.InstallCC(v, peer0Org2); err != nil {
// 		log.Panicf("Intall chaincode error: %v", err)
// 	}
// 	log.Println("Chaincode has been installed on org2's peers")
//
// 	// Upgrade chaincode only need once for each channel
// 	if err := cli1.UpgradeCC(v, peer0Org1); err != nil {
// 		log.Panicf("Upgrade chaincode error: %v", err)
// 	}
// 	log.Println("Upgrade chaincode success for channel")
//
// 	if err := cli1.InvokeCC([]string{"peer0.org1.example.com", "peer0.org2.example.com"}); err != nil {
// 		log.Panicf("Invoke chaincode error: %v", err)
// 	}
// 	log.Println("Invoke chaincode success")
//
// 	if err := cli1.QueryCC("peer0.org2.example.com", "a"); err != nil {
// 		log.Panicf("Query chaincode error: %v", err)
// 	}
// 	log.Println("Query chaincode success on peer0.org2")
// }
