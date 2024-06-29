package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing message validation
type SmartContract struct {
	contractapi.Contract
}

// Message IDs
const (
	RDSO string = "RDSO"
	IPP  string = "IPP"
)

// Message represents the structure for a message on the ledger
type GooseData struct {
	ID        string      `json:"ID"`
	Message   interface{} `json:"Message"`
	Timestamp string      `json:"Timestamp"`
	Status    string      `json:"Status"`
}

// getCurrentTimestamp returns the current transaction timestamp as a string
func getCurrentTimestamp(ctx contractapi.TransactionContextInterface) string {
	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)
}

// putState writes a value to the ledger after marshaling it to JSON
func putState(ctx contractapi.TransactionContextInterface, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(key, data)
}

// InitLedger initializes the ledger with a list of valid strings
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	IDs := []string{RDSO, IPP}

	data, err := json.Marshal(IDs)
	if err != nil {
		return fmt.Errorf("failed to marshal valid strings: %v", err)
	}

	currentTimestamp := getCurrentTimestamp(ctx)
	messages := []GooseData{
		{ID: RDSO, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "TRUE"}, Status: "Valid"},
		{ID: IPP, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "FALSE"}, Status: "Valid"},
	}

	// Initialize the messages
	for _, msg := range messages {
		msg.Timestamp = currentTimestamp
		if err := putState(ctx, msg.ID, msg); err != nil {
			return fmt.Errorf("failed to initialize goose message data: %v", err)
		}
	}

	//Intitialize the IDs
	if err := ctx.GetStub().PutState("IDs", data); err != nil {
		return fmt.Errorf("failed to initialize Ids: %v", err)
	}

	return nil
}

// Validate checks if the given id is in the set of ids
func (s *SmartContract) Validate(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState("IDs")

	if err != nil {
		return false, fmt.Errorf("failed to retrieve valid IDs from world state: %v", err)
	}

	if data == nil {
		return false, fmt.Errorf("valid IDs not found")
	}

	var IDs []string
	err = json.Unmarshal(data, &IDs)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal IDs: %v", err)
	}

	for _, str := range IDs {
		if str == id {
			return true, nil
		}
	}

	return false, nil
}

// BookKeeping adds new entires to the ledger and update the world state
func (s *SmartContract) BookKeeping(ctx contractapi.TransactionContextInterface, id string, gooseData map[string]interface{}, status string) error {
	if id != RDSO && id != IPP {
		return fmt.Errorf("invalid ID: %s", id)
	}

	var newState GooseData

	newState.ID = id
	newState.Message = gooseData
	newState.Timestamp = getCurrentTimestamp(ctx)
	newState.Status = status

	if err := putState(ctx, id, newState); err != nil {
		return fmt.Errorf("failed to update blockchain with bookkeeping data: %v", err)
	}

	return nil
}

// Read retrieves data for a specific id from the world state
func (s *SmartContract) Read(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	data, err := ctx.GetStub().GetState(id)

	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}

	if data == nil {
		return "", fmt.Errorf("%s does not exist", id)
	}

	return string(data), nil
}

// GetState retrieves all four predefined data assets from the world state.
func (s *SmartContract) GetState(ctx contractapi.TransactionContextInterface) (map[string]interface{}, error) {
	ids := []string{RDSO, IPP, "IDs"}
	allData := make(map[string]interface{})

	for _, id := range ids {
		data, err := s.Read(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to read data for %s: %v", id, err)
		}

		var dataObj interface{}
		err = json.Unmarshal([]byte(data), &dataObj)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data for %s: %v", id, err)
		}

		allData[id] = dataObj
	}

	return allData, nil
}

// GetHistory retrieves the history for a particular ID
func (s *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, ID string) ([]map[string]interface{}, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving model history: %v", err)
	}
	defer resultsIterator.Close()

	var history []map[string]interface{}
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error reading model history: %v", err)
		}

		var tx map[string]interface{}
		if err := json.Unmarshal(response.Value, &tx); err != nil {
			return nil, fmt.Errorf("error unmarshaling transaction: %v", err)
		}

		historyRecord := map[string]interface{}{
			"TxId":      response.TxId,
			"Timestamp": time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String(),
			"Value":     tx,
		}
		history = append(history, historyRecord)
	}

	return history, nil
}

// UpdateIDs Replaces the Current IDs in the World State with the New IDs
func (s *SmartContract) UpdateIDs(ctx contractapi.TransactionContextInterface, newIDs []string) error {
	data, err := json.Marshal(newIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal new IDs: %v", err)
	}

	if err := ctx.GetStub().PutState("IDs", data); err != nil {
		return fmt.Errorf("failed to update IDs: %v", err)
	}

	return nil
}
