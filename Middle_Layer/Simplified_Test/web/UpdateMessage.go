package web

import (
    "encoding/json"
    "fmt"
    "net/http"
    //"time"
)

// UpdateMessage godoc
// @Summary Update a message on the ledger
// @Description Update or add a message by its ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id formData string true "ID of the Message"
// @Param messageType formData string true "Type of the Message (Standard or Corrective)"
// @Param messageContent formData string true "Content of the Message as a JSON string"
// @Success 200 {string} string "Message Updated"
// @Failure 400 {string} string "Error occurred"
// @Router /update [post]

func (setup *OrgSetup) UpdateMessage(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received UpdateMessage request")

    var requestData struct {
        ID             string                 `json:"id"`
        MessageType    string                 `json:"messageType"`
        MessageContent map[string]interface{} `json:"messageContent"`
    }

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := NewWrappedContract(network.GetContract(setup.Chaincode))

    messageContentBytes, err := json.Marshal(requestData.MessageContent)
    if err != nil {
        http.Error(w, "JSON Marshal error: "+err.Error(), http.StatusBadRequest)
        return
    }

    result, err := contract.SubmitTransactionWithTiming("UpdateMessage", requestData.ID, string(messageContentBytes), requestData.MessageType)
    if err != nil {
        http.Error(w, "Error invoking UpdateMessage: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Message Updated!\n %s", result)
}

// func (setup *OrgSetup) UpdateMessage(w http.ResponseWriter, r *http.Request) {
//     fmt.Println("Received UpdateMessage request")
    
//     // Start timing
//     start := time.Now()
    
//     var requestData struct {
//         ID             string                   `json:"id"`
//         MessageType    string                   `json:"messageType"`
//         MessageContent map[string]interface{}   `json:"messageContent"`
//     }

//     if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
//         http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
//         return
//     }

//     network := setup.Gateway.GetNetwork(setup.Channel)
//     contract := network.GetContract(setup.Chaincode)

//     messageContentBytes, err := json.Marshal(requestData.MessageContent)
//     if err != nil {
//         http.Error(w, "JSON Marshal error: "+err.Error(), http.StatusBadRequest)
//         return
//     }

//     result, err := contract.SubmitTransaction("UpdateMessage", requestData.ID, string(messageContentBytes), requestData.MessageType)
//     if err != nil {
//         http.Error(w, "Error invoking UpdateMessage: "+err.Error(), http.StatusInternalServerError)
//         return
//     }

//     // Calculate time spent
//     end := time.Now()
//     timeSpent := end.Sub(start).Seconds()

//     // Broadcast the time spent
//     broadcastTimeSpent(timeSpent)

//     fmt.Fprintf(w, "Message Updated!\n %s", result)
// }
