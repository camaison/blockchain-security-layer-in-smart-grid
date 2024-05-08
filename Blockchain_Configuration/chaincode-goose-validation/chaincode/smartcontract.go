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

// StatusType represents the validation status
type StatusType string

const (
	// Message types
	Standard   MessageType = "Standard"
	Corrective MessageType = "Corrective"

	// Status types
	Valid   StatusType = "Valid"
	Invalid StatusType = "Invalid"

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
	Status     StatusType  `json:"Status"`
}

// InitLedger initializes the ledger with sample messages and responses
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	currentTimestamp := getCurrentTimestamp(ctx)
	messages := []Message{
		{ID: RDSO_PubMessage, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "TRUE"}, Type: Standard},
		{ID: IPP_PubMessage, Message: map[string]interface{}{"t": currentTimestamp, "stNum": 0, "allData": "FALSE"}, Type: Standard},
	}

	responses := []Response{
		{ID: RDSO_ValidationMessage, Subscribed: map[string]interface{}{}, Published: map[string]interface{}{}, Status: Valid},
		{ID: IPP_ValidationMessage, Subscribed: map[string]interface{}{}, Published: map[string]interface{}{}, Status: Valid},
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
func (s *SmartContract) UpdateMessage(ctx contractapi.TransactionContextInterface, id string, messageContent interface{}, messageType MessageType) error {
	if id != RDSO_PubMessage && id != IPP_PubMessage {
		return fmt.Errorf("invalid message ID: %s", id)
	}

	// Retrieve the existing message if it exists
	var message Message
	message, err := s.ReadData(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read existing message: %v", err)
	}

	if message == nil {
		message.ID = id
		message.Type = Standard // Default type
	}

	message.Message = messageContent
	message.Timestamp = getCurrentTimestamp(ctx)
	message.Type = messageType

	return putState(ctx, id, message)
}

// RespondToMessage updates or creates a response in the ledger
func (s *SmartContract) RespondToMessage(ctx contractapi.TransactionContextInterface, id string, subscribedContent, publishedContent interface{}) (string, error) {
	if id != RDSO_ValidationMessage && id != IPP_ValidationMessage {
		return "", fmt.Errorf("invalid response ID: %s", id)
	}

    // Update the corresponding message first with the publishedContent
	messageIDToUpdate := RDSO_PubMessage
	if id == IPP_ValidationMessage {
		messageIDToUpdate = IPP_PubMessage
	}

    if err := s.UpdateMessage(ctx, messageIDToUpdate, publishedContent, Standard); err != nil {
		return "", fmt.Errorf("failed to update message: %v", err)
	}

	var response Response
	response, err := s.ReadData(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to read existing response: %v", err)
	}

	if response == nil {
		response.ID = id
		response.Status = Valid // Default status
	}

	// Update the response content, timestamp, and status
	response.Subscribed = subscribedContent
	response.Published = publishedContent
	response.Timestamp = getCurrentTimestamp(ctx)

    
	// Determine the ID of the message to validate against
	validateAgainstID := RDSO_PubMessage
	if id == RDSO_ValidationMessage {
		validateAgainstID = IPP_PubMessage
	}

    // Perform the validation
	isValid, err := s.ValidateMessage(ctx, validateAgainstID, subscribedContent)
	if err != nil {
		return "", fmt.Errorf("validation failed: %v", err)
	}

	if isValid {
		response.Status = Valid
	} else {
		response.Status = Invalid
	}

    if err := putState(ctx, id, response); err != nil {
		return "", fmt.Errorf("failed to put updated response: %v", err)
	}

	// Return the validation result
	return string(response.Status), nil
}

// ValidateMessage compares the subscribed content with the message content in the world state.
func (s *SmartContract) ValidateMessage(ctx contractapi.TransactionContextInterface, messageID string, subscribedContent interface{}) (bool, error) {
	var message Message
	message, err := s.ReadData(ctx, messageID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch message for validation: %v", err)
	}

	if message == nil {
		return false, fmt.Errorf("message %s does not exist for validation", messageID)
	}

	subscribedBytes, err := json.Marshal(subscribedContent)
	if err != nil {
		return false, fmt.Errorf("failed to marshal subscribed content: %v", err)
	}

	messageBytes, err := json.Marshal(message.Message)
	if err != nil {
		return false, fmt.Errorf("failed to marshal message content: %v", err)
	}

	// Compare the entire JSON strings
	return string(subscribedBytes) == string(messageBytes), nil
}

// ReadData retrieves a specific state from the ledger
func (s *SmartContract) ReadData(ctx contractapi.TransactionContextInterface, id string) (interface{}, error) {
	dataJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if dataJSON == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

    var data interface{}
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return data, nil
}

// GetAllData retrieves all four predefined data assets from the world state.
func (s *SmartContract) GetAllData(ctx contractapi.TransactionContextInterface) (map[string]interface{}, error) {
	ids := []string{RDSO_PubMessage, IPP_PubMessage, RDSO_ValidationMessage, IPP_ValidationMessage}
	allData := make(map[string]interface{})

	for _, id := range ids {
		var data interface{}
		data, err := s.ReadData(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to read data for %s: %v", id, err)
		}
		if data != nil {
			allData[id] = data
		}
	}

	return allData, nil
}

// GetHistoryForID retrieves the history for a particular ID
func (s *SmartContract) GetHistoryForID(ctx contractapi.TransactionContextInterface, id string) ([]map[string]interface{}, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for %s: %v", id, err)
	}
	defer resultsIterator.Close()

	var history []map[string]interface{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read next item in history: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(queryResponse.Value, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal history data: %v", err)
		}
		history = append(history, data)
	}
	return history, nil
}

