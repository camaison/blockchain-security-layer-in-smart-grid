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
func (s *SmartContract) UpdateMessage(ctx contractapi.TransactionContextInterface, id string, messageContent map[string]interface{}, messageTypeStr string) error {
    // Convert string to MessageType
    messageType := MessageType(messageTypeStr)

	// Validate the message ID
    if id != RDSO_PubMessage && id != IPP_PubMessage {
        return fmt.Errorf("invalid message ID: %s", id)
    }

    // Validate the message type
    if messageType != Standard && messageType != Corrective {
        return fmt.Errorf("invalid message type: %s", messageType)
    }

    // Retrieve the existing message from the ledger
    exists, err := s.ReadData(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to read existing message: %v", err)
    }

    // Cast the existing data to a Message structure
    var message Message
    switch ex := exists.(type) {
    case map[string]interface{}:
        // Manually mapping fields to ensure structure matches
        message.ID = ex["ID"].(string)
        message.Timestamp = ex["Timestamp"].(string)
        message.Type = MessageType(ex["Type"].(string))
        if msg, ok := ex["Message"].(map[string]interface{}); ok {
            message.Message = msg
        } else {
            return fmt.Errorf("existing message content has incorrect format")
        }
    default:
        return fmt.Errorf("unexpected format of existing message data")
    }

    // Update the message content
    message.Message = messageContent
    message.Timestamp = getCurrentTimestamp(ctx)
    message.Type = messageType

    // Store the updated message back to the ledger
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

    // Determine the message ID to update based on the response ID
    messageIDToUpdate := RDSO_PubMessage
    if id == IPP_ValidationMessage {
        messageIDToUpdate = IPP_PubMessage
    }

    // Update the corresponding message with the published content
    if err := s.UpdateMessage(ctx, messageIDToUpdate, publishedContent, "Standard"); err != nil {
        return "", fmt.Errorf("failed to update message: %v", err)
    }

    // Fetch the existing response data or initialize if it does not exist
    existingData, err := s.ReadData(ctx, id)
    if err != nil {
        return "", fmt.Errorf("failed to read existing response: %v", err)
    }

    // Decode or initialize the response structure
    var response Response
    if existingData != nil {
        switch ed := existingData.(type) {
        case map[string]interface{}:
            tmpBytes, _ := json.Marshal(ed)
            if err := json.Unmarshal(tmpBytes, &response); err != nil {
                return "", fmt.Errorf("failed to unmarshal response: %v", err)
            }
        default:
            return "", fmt.Errorf("unexpected format of existing response data")
        }
    } else {
        response = Response{
            ID: id,
            Status: Valid, // Default status if not existing
        }
    }

    // Update the response content and other fields
    response.Subscribed = subscribedContent
    response.Published = publishedContent
    response.Timestamp = getCurrentTimestamp(ctx)

    // Validate the subscribed content against the opposite message content
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

    // Store the updated response
    if err := putState(ctx, id, response); err != nil {
        return "", fmt.Errorf("failed to put updated response: %v", err)
    }

    return string(response.Status), nil
}

// ValidateMessage compares the subscribed content with the message content in the world state.
func (s *SmartContract) ValidateMessage(ctx contractapi.TransactionContextInterface, messageID string, subscribedContent map[string]interface{}) (bool, error) {
    messageData, err := s.ReadData(ctx, messageID)
    if err != nil {
        return false, fmt.Errorf("failed to fetch message for validation: %v", err)
    }

    var message Message
    switch md := messageData.(type) {
    case map[string]interface{}:
        tmpBytes, _ := json.Marshal(md)
        if err := json.Unmarshal(tmpBytes, &message); err != nil {
            return false, fmt.Errorf("failed to unmarshal message: %v", err)
        }
    default:
        return false, fmt.Errorf("unexpected format of message data")
    }

    // Convert the Message part to map[string]interface{} if needed
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
            // Perform a simple equality check, this can be adjusted to be a deep equality check if needed
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