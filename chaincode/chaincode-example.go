package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type KVStore struct {
}

func (st *KVStore) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("chaincode init success"))
}

func (st *KVStore) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fname, params := stub.GetFunctionAndParameters()
	if fname == "set" {
		return set(stub, params[0], params[1])
	} else if fname == "get" {
		return get(stub, params[0])
	}
	return shim.Error(fmt.Sprintf("Invalid Invoke function: %s", fname))
}

func set(stub shim.ChaincodeStubInterface, key string, value string) pb.Response {
	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("set success"))
}

func get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	// If the key does not exist in the state database, (nil, nil) is returned.
	if data == nil {
		return shim.Error("get failed for data is nil")
	}
	return shim.Success(data)
}

func main() {
	if err := shim.Start(new(KVStore)); err != nil {
		log.Printf("chaincode start failed: %s", err)
		return
	}
	log.Printf("chaincode start success")
}
