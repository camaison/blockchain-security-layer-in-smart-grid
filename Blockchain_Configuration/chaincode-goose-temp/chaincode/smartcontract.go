package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing message validation
type SmartContract struct {
	contractapi.Contract
}

// InitLedger initializes the ledger with a list of valid strings
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	validStrings := []string{"RDSO", "IPP"}

	data, err := json.Marshal(validStrings)
	if err != nil {
		return fmt.Errorf("failed to marshal valid strings: %v", err)
	}

	return ctx.GetStub().PutState("validStrings", data)
}

// Validate checks if the given value is in the set of valid strings
func (s *SmartContract) Validate(ctx contractapi.TransactionContextInterface, value string) (bool, error) {
	data, err := ctx.GetStub().GetState("validStrings")
	if err != nil {
		return false, fmt.Errorf("failed to read valid strings from world state: %v", err)
	}
	if data == nil {
		return false, fmt.Errorf("valid strings not found")
	}

	var validStrings []string
	err = json.Unmarshal(data, &validStrings)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal valid strings: %v", err)
	}

	for _, str := range validStrings {
		if str == value {
			return true, nil
		}
	}

	return false, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
