package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) UpdateIDs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received UpdateId request")

	var requestData struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	messageContentBytes, err := json.Marshal(requestData.IDs)
	if err != nil {
		http.Error(w, "JSON Marshal error: "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := contract.SubmitTransaction("UpdateIDs", string(messageContentBytes))
	if err != nil {
		http.Error(w, "Error invoking UpdateIDs Function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Validation IDs Updated to %s\n %s", requestData.IDs, result)
}
