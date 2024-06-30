package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) ValidateID(w http.ResponseWriter, r *http.Request) {
	// Define a structure for the expected JSON payload
	var requestData struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received Validate request for id:", requestData.ID)

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Submit transaction to the ledger to validate the value
	result, err := contract.EvaluateTransaction("Validate", requestData.ID)
	if err != nil {
		http.Error(w, "Error invoking Validate function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Assume the result is a simple boolean as a string; adjust as necessary
	isValid := string(result) == "true"

	// Send the response with validation result
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(map[string]bool{isValid})
	json.NewEncoder(w).Encode(map[string]bool{"isValid": isValid})

}
