package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %s", err)
		return
	}

	id := r.FormValue("id")
	messageTypeStr := r.FormValue("messageType")

	// Parse the messageContent from a JSON string in the form data
	messageContent := make(map[string]interface{})
	if err := json.Unmarshal([]byte(r.FormValue("messageContent")), &messageContent); err != nil {
		fmt.Fprintf(w, "JSON Unmarshal error: %s", err)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Convert messageContent to a JSON string for chaincode interaction
	messageContentBytes, err := json.Marshal(messageContent)
	if err != nil {
		fmt.Fprintf(w, "JSON Marshal error: %s", err)
		return
	}

	// Submit transaction to the ledger
	result, err := contract.SubmitTransaction("UpdateMessage", id, string(messageContentBytes), messageTypeStr)
	if err != nil {
		fmt.Fprintf(w, "Error invoking UpdateMessage: %s", err)
		return
	}
	fmt.Fprintf(w, "Message Updated: %s", result)
}
