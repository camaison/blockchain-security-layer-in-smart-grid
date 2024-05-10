package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing message and response records
type SmartContract struct {
	contractapi.Contract
}

// MessageType represents the type of a Message
type MessageType string

// ResultType represents the validation Result
type ResultType string

const (
	// Message types
	Standard   MessageType = "Standard"
	Corrective MessageType = "Corrective"

	// Result types
	Valid   ResultType = "Valid"
	Invalid ResultType = "Invalid"

	// Message IDs
	RDSO_PubMessage string = "RDSO_PubMessage"
	IPP_PubMessage  string = "IPP_PubMessage"

	// Response IDs
	RDSO_ValidationMessage string = "RDSO_ValidationMessage"
	IPP_ValidationMessage  string = "IPP_ValidationMessage"
)

// Message represents the structure for a message on the ledger
type Message struct {
	ID        string      `json:"ID"`
	Message   interface{} `json:"Message"`
	Timestamp string      `json:"Timestamp"`
	Type      MessageType `json:"Type"`
}

// Response represents the structure for a response on the ledger
type Response struct {
	ID         string      `json:"ID"`
	Subscribed interface{} `json:"Subscribed"`
	Published  interface{} `json:"Published"`
	Timestamp  string      `json:"Timestamp"`
	Result     ResultType  `json:"Result"`
}

// InitLedger initializes the ledger with sample messages and responses
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	currentTimestamp := getCurrentTimestamp(ctx)
	messages := []Message{
		{ID: RDSO_PubMessage, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "TRUE"}, Type: Standard},
		{ID: IPP_PubMessage, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "FALSE"}, Type: Standard},
	}

	responses := []Response{
		{ID: RDSO_ValidationMessage, Subscribed: map[string]interface{}{}, Published: map[string]interface{}{}, Result: Valid},
		{ID: IPP_ValidationMessage, Subscribed: map[string]interface{}{}, Published: map[string]interface{}{}, Result: Valid},
	}


	for _, msg := range messages {
		msg.Timestamp = currentTimestamp
		if err := putState(ctx, msg.ID, msg); err != nil {
			return fmt.Errorf("failed to initialize message: %v", err)
		}
	}

	for _, resp := range responses {
		resp.Timestamp = currentTimestamp
		if err := putState(ctx, resp.ID, resp); err != nil {
			return fmt.Errorf("failed to initialize response: %v", err)
		}
	}

	return nil
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

// UpdateMessage updates or adds a message in the ledger
func (s *SmartContract) UpdateMessage(ctx contractapi.TransactionContextInterface, id string, messageContent map[string]interface{}, messageTypeStr string) error {
    messageType := MessageType(messageTypeStr)

    if id != RDSO_PubMessage && id != IPP_PubMessage {
        return fmt.Errorf("invalid message ID: %s", id)
    }

    if messageType != Standard && messageType != Corrective {
        return fmt.Errorf("invalid message type: %s", messageType)
    }

    exists, err := s.ReadData(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to read existing message: %v", err)
    }

    var message Message
    err = json.Unmarshal([]byte(exists), &message)
    if err != nil {
        return fmt.Errorf("failed to unmarshal existing message: %v", err)
    }

    message.Message = messageContent
    message.Timestamp = getCurrentTimestamp(ctx)
    message.Type = messageType

    if err := putState(ctx, id, message); err != nil {
        return fmt.Errorf("failed to put updated message: %v", err)
    }

    return nil
}


// RespondToMessage updates or creates a response in the ledger
func (s *SmartContract) RespondToMessage(ctx contractapi.TransactionContextInterface, id string, subscribedContent, publishedContent map[string]interface{}) (string, error) {
    if id != RDSO_ValidationMessage && id != IPP_ValidationMessage {
        return "", fmt.Errorf("invalid response ID: %s", id)
    }

    messageIDToUpdate := RDSO_PubMessage
    if id == IPP_ValidationMessage {
        messageIDToUpdate = IPP_PubMessage
    }

    if err := s.UpdateMessage(ctx, messageIDToUpdate, publishedContent, "Standard"); err != nil {
        return "", fmt.Errorf("failed to update message: %v", err)
    }

    existingData, err := s.ReadData(ctx, id)
    if err != nil {
        return "", fmt.Errorf("failed to read existing response: %v", err)
    }

    var response Response
    err = json.Unmarshal([]byte(existingData), &response)
    if err != nil {
        return "", fmt.Errorf("failed to unmarshal existing response: %v", err)
    }

    response.Subscribed = subscribedContent
    response.Published = publishedContent
    response.Timestamp = getCurrentTimestamp(ctx)

    validateAgainstID := RDSO_PubMessage
    if id == RDSO_ValidationMessage {
        validateAgainstID = IPP_PubMessage
    }

    isValid, err := s.ValidateMessage(ctx, validateAgainstID, subscribedContent)
    if err != nil {
        return "", fmt.Errorf("validation failed: %v", err)
    }

    if isValid {
        response.Result = Valid
    } else {
        response.Result = Invalid
    }

    if err := putState(ctx, id, response); err != nil {
        return "", fmt.Errorf("failed to put updated response: %v", err)
    }

    return string(response.Result), nil
}

// ValidateMessage compares the subscribed content with the message content in the world state.
func (s *SmartContract) ValidateMessage(ctx contractapi.TransactionContextInterface, messageID string, subscribedContent map[string]interface{}) (bool, error) {
    messageDataStr, err := s.ReadData(ctx, messageID)
    if err != nil {
        return false, fmt.Errorf("failed to fetch message for validation: %v", err)
    }

    // Unmarshal the JSON string into Message struct
    var message Message
    if err := json.Unmarshal([]byte(messageDataStr), &message); err != nil {
        return false, fmt.Errorf("failed to unmarshal message: %v", err)
    }

    // Access the Message part which is expected to be a map[string]interface{}
    messageContent, ok := message.Message.(map[string]interface{})
    if !ok {
        return false, fmt.Errorf("message content is not in the expected format")
    }

    // Compare the content directly
    if len(messageContent) != len(subscribedContent) {
        return false, nil
    }
    
    for key, msgVal := range messageContent {
        if subVal, exists := subscribedContent[key]; exists {
            // Perform a simple equality check
            if fmt.Sprintf("%v", msgVal) != fmt.Sprintf("%v", subVal) {
                return false, nil
            }
        } else {
            return false, nil
        }
    }

    return true, nil
}


// ReadData retrieves a specific state from the ledger
func (s *SmartContract) ReadData(ctx contractapi.TransactionContextInterface, id string) (string, error) {
    data, err := ctx.GetStub().GetState(id)
    if err != nil {
        return "", fmt.Errorf("failed to read from world state: %v", err)
    }
    if data == nil {
        return "", fmt.Errorf("%s does not exist", id)
    }

    var dataMap map[string]interface{}
    if err := json.Unmarshal(data, &dataMap); err != nil {
        return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
    }

    jsonData, err := json.Marshal(dataMap)
    if err != nil {
        return "", fmt.Errorf("failed to marshal JSON: %v", err)
    }

    return string(jsonData), nil
}


// GetAllData retrieves all four predefined data assets from the world state.
func (s *SmartContract) GetAllData(ctx contractapi.TransactionContextInterface) (map[string]interface{}, error) {
    ids := []string{RDSO_PubMessage, IPP_PubMessage, RDSO_ValidationMessage, IPP_ValidationMessage}
    allData := make(map[string]interface{})

    for _, id := range ids {
        data, err := s.ReadData(ctx, id)
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

// GetTxnHistory retrieves the history for a particular ID
func (s *SmartContract) GetTxnHistory(ctx contractapi.TransactionContextInterface, ID string) ([]map[string]interface{}, error) {
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