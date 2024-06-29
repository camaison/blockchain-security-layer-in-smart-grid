package main

import (
	"log"

	"goose-temp-chaincode/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	messageChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating messagevalidation chaincode: %s", err.Error())
	}

	if err := messageChaincode.Start(); err != nil {
		log.Panicf("Error starting messagevalidation chaincode: %s", err.Error())
	}
}
