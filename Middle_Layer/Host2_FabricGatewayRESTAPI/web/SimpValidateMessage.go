package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SimpValidateMessage godoc
// @Summary Validate a message in the ledger
// @Description Compares the given value with the valid strings in the world state
// @Tags validation
// @Accept json
// @Produce json
// @Param value formData string true "SimpValue to validate"
// @Success 200 {object} map[string]bool "Simple Validation Result"
// @Failure 400 {string} string "Error occurred"
// @Router /simpvalidate [post]
func (setup *OrgSetup) SimpValidateMessage(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received SimpValidateMessage request")

    // Define a structure for the expected JSON payload
    type Request struct {
        Value string `json:"value"`
    }

    var requestData Request
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract("goose-temp")

    // Submit transaction to the ledger to validate the value
    result, err := contract.EvaluateTransaction("Validate", requestData.Value)
    if err != nil {
        http.Error(w, "Error invoking Validate: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Assume the result is a simple boolean as a string; adjust as necessary
    isValid := string(result) == "true"

    // Send the response with validation result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{isValid})
}