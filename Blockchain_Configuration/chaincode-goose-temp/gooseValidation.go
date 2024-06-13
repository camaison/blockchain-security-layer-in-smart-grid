package main

import (
    "log"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
    "goose-temp-chaincode/chaincode"
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
