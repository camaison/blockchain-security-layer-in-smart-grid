package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ValidateMessage godoc
// @Summary Validate a message in the ledger
// @Description Compares the subscribed content with the message content in the world state
// @Tags validation
// @Accept json
// @Produce json
// @Param messageID formData string true "ID of the Message to validate"
// @Param subscribedContent formData string true "Subscribed Content as a JSON string"
// @Success 200 {object} map[string]interface{} "Validation Result"
// @Failure 400 {string} string "Error occurred"
// @Router /validate [post]
func (setup *OrgSetup) ValidateMessage(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received ValidateMessage request")

    // Define a structure for the expected JSON payload
    type Request struct {
        MessageID         string                 `json:"messageID"`
        SubscribedContent map[string]interface{} `json:"subscribedContent"`
    }

    var requestData Request
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    // Convert subscribedContent to a JSON string for chaincode interaction
    subscribedContentBytes, err := json.Marshal(requestData.SubscribedContent)
    if err != nil {
        http.Error(w, "JSON Marshal error for subscribedContent: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Submit transaction to the ledger to validate the message
    result, err := contract.EvaluateTransaction("ValidateMessage", requestData.MessageID, string(subscribedContentBytes))
    if err != nil {
        http.Error(w, "Error invoking ValidateMessage: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Assume the result is a simple boolean as a string; adjust as necessary
    isValid := string(result) == "true"

    // Send the response with validation result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"isValid": isValid})
}