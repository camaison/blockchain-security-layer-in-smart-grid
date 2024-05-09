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
// @Router /validateMessage [post]
func (setup *OrgSetup) ValidateMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received ValidateMessage request")
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %s", err)
		return
	}

	messageID := r.FormValue("messageID")

	// Parse the subscribedContent from JSON string in the form data
	subscribedContent := make(map[string]interface{})
	if err := json.Unmarshal([]byte(r.FormValue("subscribedContent")), &subscribedContent); err != nil {
		fmt.Fprintf(w, "JSON Unmarshal error for subscribedContent: %s", err)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Convert subscribedContent to a JSON string for chaincode interaction
	subscribedContentBytes, err := json.Marshal(subscribedContent)
	if err != nil {
		fmt.Fprintf(w, "JSON Marshal error for subscribedContent: %s", err)
		return
	}

	// Submit transaction to the ledger to validate the message
	result, err := contract.EvaluateTransaction("ValidateMessage", messageID, string(subscribedContentBytes))
	if err != nil {
		fmt.Fprintf(w, "Error invoking ValidateMessage: %s", err)
		return
	}

	// Convert the result (true/false) into a map to send as JSON
	validationResult := map[string]interface{}{
		"isValid": string(result) == "true",
	}

	// Send the response with validation result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validationResult)
}
