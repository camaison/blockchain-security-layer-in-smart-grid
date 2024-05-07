package chaincode

import (
    "encoding/json"
	"fmt"
    "time"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for controlling the message exchange
type SmartContract struct {
    contractapi.Contract
}

// Message represents the structure for a message on the ledger
type Message struct {
    PartyID   string `json:"partyID"`
    Content   string `json:"content"`
    Timestamp string `json:"timestamp"`
    Type      string `json:"type"` // "Standard" or "Corrective"
}

// Response captures the response and validation status
type Response struct {
    PartyID   		 string `json:"partyID"`
	ReceivedContent  string `json:"receivedContent"`
    ResponseContent  string `json:"responseContent"`
    Timestamp        string `json:"timestamp"`
    Status           string `json:"status"`
}

// Init function to initialize the ledger with sample data
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
    txTimestamp, err := ctx.GetStub().GetTxTimestamp()
    if err != nil {
        return fmt.Errorf("failed to get transaction timestamp: %v", err)
    }
    timestamp := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).String()

    messages := []Message{
        {PartyID: "party1", Content: "Initial message from party1", Timestamp: timestamp, Type: "Standard"},
        {PartyID: "party2", Content: "Initial message from party2", Timestamp: timestamp, Type: "Standard"},
    }

    responses := []Response{
        {PartyID: "party1", ReceivedContent: "Initial message from party1", ResponseContent: "Initial response from party2", Timestamp: timestamp, Status: "Valid"},
        {PartyID: "party2", ReceivedContent: "Initial message from party2", ResponseContent: "Initial response from party1", Timestamp: timestamp, Status: "Valid"},
    }

    for i, message := range messages {
        messageJSON, _ := json.Marshal(message)
        if err := ctx.GetStub().PutState("Message"+fmt.Sprint(i), messageJSON); err != nil {
            return fmt.Errorf("failed to put to world state. %v", err)
        }
    }

    for i, response := range responses {
        responseJSON, _ := json.Marshal(response)
        if err := ctx.GetStub().PutState("Response"+fmt.Sprint(i), responseJSON); err != nil {
            return fmt.Errorf("failed to put to world state. %v", err)
        }
    }

    return nil
}

// UpdateMessage adds a new message to the ledger
func (s *SmartContract) UpdateMessage(ctx contractapi.TransactionContextInterface, partyID string, messageContent string, messageType string) error {
    txTimestamp, err := ctx.GetStub().GetTxTimestamp()
    if err != nil {
        return fmt.Errorf("failed to get transaction timestamp: %v", err)
    }
    timestamp := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).String()

    message := Message{
        PartyID:    partyID,
        Content:    messageContent,
        Timestamp:  timestamp,
        Type:       messageType,
    }
    messageJSON, _ := json.Marshal(message)
    return ctx.GetStub().PutState(partyID + "_updateMessage", messageJSON)
}


// RespondToMessage handles responses to messages
func (s *SmartContract) RespondToMessage(ctx contractapi.TransactionContextInterface, receivedMessages map[string]string) (string, error) {
    receivedContent := receivedMessages["receivedContent"]
    responseContent := receivedMessages["responseContent"]
    senderPartyID := receivedMessages["senderPartyID"]
    responderPartyID := receivedMessages["responderPartyID"]

    s.UpdateMessage(ctx, responderPartyID+"_updateMessage", responseContent, "Standard")
    validationStatus, err := s.ValidateMessage(ctx, senderPartyID+"_updateMessage", receivedContent)
    if err != nil {
        return "", err
    }
    s.LogResponse(ctx, responderPartyID+"_responseMessage", receivedContent, responseContent, validationStatus)
    return validationStatus, nil
}

// ValidateMessage checks if a message content matches the expected content
func (s *SmartContract) ValidateMessage(ctx contractapi.TransactionContextInterface, partyID string, receivedContent string) (string, error) {
    messageJSON, err := ctx.GetStub().GetState(partyID)
    if err != nil {
        return "", err
    }

    var message Message
    json.Unmarshal(messageJSON, &message)
    if message.Content == receivedContent {
        return "Valid", nil
    } else {
        return "Invalid", nil
    }
}

// LogResponse logs the response and the validation result
func (s *SmartContract) LogResponse(ctx contractapi.TransactionContextInterface, responseID string, receivedContent string, responseContent string, status string) error {
    response := Response{
        ReceivedContent: receivedContent,
        ResponseContent: responseContent,
        Timestamp:       time.Now().String(),
        Status:          status,
    }
    responseJSON, _ := json.Marshal(response)
    return ctx.GetStub().PutState(responseID, responseJSON)
}

// ReadCurrentData retrieves a specific state from the ledger using ID
func (s *SmartContract) ReadCurrentData(ctx contractapi.TransactionContextInterface, id string) (*Message, error) {
    messageJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if messageJSON == nil {
        return nil, fmt.Errorf("the message %s does not exist", id)
    }

    var message Message
    err = json.Unmarshal(messageJSON, &message)
    if err != nil {
        return nil, err
    }
    return &message, nil
}

// GetAllCurrentData retrieves all messages from the ledger
func (s *SmartContract) GetAllCurrentData(ctx contractapi.TransactionContextInterface) ([]Message, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var messages []Message
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var message Message
        err = json.Unmarshal(queryResponse.Value, &message)
        if err != nil {
            return nil, err
        }
        messages = append(messages, message)
    }
    return messages, nil
}

// GetHistoryForID retrieves the history for a particular ID
func (s *SmartContract) GetHistoryForID(ctx contractapi.TransactionContextInterface, id string) ([]Response, error) {
    resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var responses []Response
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var response Response
        err = json.Unmarshal(queryResponse.Value, &response)
        if err != nil {
            return nil, err
        }
        responses = append(responses, response)
    }
    return responses, nil
}

// QueryByTypeOrStatus queries for messages or responses by type or status
func (s *SmartContract) QueryByTypeOrStatus(ctx contractapi.TransactionContextInterface, indexName string, attribute string) ([]Response, error) {
    queryString := fmt.Sprintf(`{"selector":{"%s":"%s"}}`, indexName, attribute)
    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var responses []Response
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var response Response
        err = json.Unmarshal(queryResponse.Value, &response)
        if err != nil {
            return nil, err
        }
        responses = append(responses, response)
    }
    return responses, nil
}