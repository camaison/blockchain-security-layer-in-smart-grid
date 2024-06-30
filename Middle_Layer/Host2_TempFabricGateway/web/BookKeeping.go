package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (setup *OrgSetup) BookKeeping(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		ID      string                 `json:"id"`
		Message map[string]interface{} `json:"message"`
		Status  string                 `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received BookKeeping request for", requestData.ID)

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := NewWrappedContract(network.GetContract(setup.Chaincode))

	messageContentBytes, err := json.Marshal(requestData.Message)
	if err != nil {
		http.Error(w, "JSON Marshal error: "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := contract.SubmitTransactionWithTiming("BookKeeping", requestData.ID, string(messageContentBytes), requestData.Status)
	if err != nil {
		http.Error(w, "Error invoking BookKeeping Function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "BookKeeping Completed Successfuly!\n %s", result)
}
