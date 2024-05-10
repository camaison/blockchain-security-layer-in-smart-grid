package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetTxnHistory godoc
// @Summary Get transaction history for a specific ID
// @Description Retrieves the transaction history for a specific ledger entry by ID
// @Tags history
// @Accept json
// @Produce json
// @Param id query string true "ID of the Data to retrieve history for"
// @Success 200 {array} map[string]interface{} "Transaction History Retrieved"
// @Failure 400 {string} string "Error occurred"
// @Router /history [get]
func (setup *OrgSetup) GetTxnHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GetTxnHistory request")

	// Extract 'id' from query parameters
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' is missing", http.StatusBadRequest)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Evaluate transaction using the GetTxnHistory function from chaincode
	result, err := contract.EvaluateTransaction("GetTxnHistory", id)
	if err != nil {
		fmt.Fprintf(w, "Error querying GetTxnHistory: %s", err)
		return
	}

	// Convert result into a JSON format that can be sent back to the client
	var history []map[string]interface{}
	if err := json.Unmarshal(result, &history); err != nil {
		fmt.Fprintf(w, "Error unmarshaling JSON data: %s", err)
		return
	}

	// Send the response with transaction history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
