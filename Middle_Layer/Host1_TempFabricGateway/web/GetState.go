package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) GetState(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GetState request")

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Call the GetState function from chaincode
	result, err := contract.EvaluateTransaction("GetState")
	if err != nil {
		fmt.Fprintf(w, "Error querying GetState Function: %s", err)
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
