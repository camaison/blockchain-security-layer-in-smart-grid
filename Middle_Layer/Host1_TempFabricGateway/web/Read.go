package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) Read(w http.ResponseWriter, r *http.Request) {
	// Extract 'id' from query parameters
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' is missing", http.StatusBadRequest)
		return
	}

	fmt.Println("Received Read request for id:", id)

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Call the Read function from chaincode
	result, err := contract.EvaluateTransaction("Read", id)
	if err != nil {
		fmt.Fprintf(w, "Error querying Read for %s: %s", id, err)
		return
	}

	// Convert result into a JSON format that can be sent back to the client
	var data interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		fmt.Fprintf(w, "Error unmarshaling JSON data: %s", err)
		return
	}

	// Send the response with data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
