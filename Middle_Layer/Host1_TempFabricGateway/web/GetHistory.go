package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) GetHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GetHistory request")

	// Extract 'id' from query parameters
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' is missing", http.StatusBadRequest)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Call the GetHistory function from chaincode
	result, err := contract.EvaluateTransaction("GetHistory", id)
	if err != nil {
		fmt.Fprintf(w, "Error querying GetHistory Function: %s", err)
		return
	}

	// Convert result into a JSON format that can be sent back to the client
	var history interface{}
	if err := json.Unmarshal(result, &history); err != nil {
		fmt.Fprintf(w, "Error unmarshaling JSON data: %s", err)
		return
	}

	// Send the response with transaction history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
