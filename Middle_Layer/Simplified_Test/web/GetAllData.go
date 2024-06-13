package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAllData godoc
// @Summary Get all predefined data assets
// @Description Retrieves all predefined data assets from the ledger
// @Tags assets
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "All Data Retrieved"
// @Failure 400 {string} string "Error occurred"
// @Router /getAll [get]
func (setup *OrgSetup) GetAllData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GetAllData request")

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Evaluate transaction using the GetAllData function from chaincode
	result, err := contract.EvaluateTransaction("GetAllData")
	if err != nil {
		fmt.Fprintf(w, "Error querying GetAllData: %s", err)
		return
	}

	// Prepare the response to return JSON data
	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		fmt.Fprintf(w, "Error unmarshaling JSON data: %s", err)
		return
	}

	// Send the response with data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
