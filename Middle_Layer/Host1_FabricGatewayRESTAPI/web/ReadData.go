package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ReadData godoc
// @Summary Read specific data from the ledger
// @Description Retrieves specific data from the ledger by ID
// @Tags data
// @Accept json
// @Produce json
// @Param id query string true "ID of the Data to retrieve"
// @Success 200 {object} interface{} "Data Retrieved"
// @Failure 400 {string} string "Error occurred"
// @Router /readData [get]
func (setup *OrgSetup) ReadData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received ReadData request")

	// Extract 'id' from query parameters
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' is missing", http.StatusBadRequest)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Evaluate transaction using the ReadData function from chaincode
	result, err := contract.EvaluateTransaction("ReadData", id)
	if err != nil {
		fmt.Fprintf(w, "Error querying ReadData: %s", err)
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
