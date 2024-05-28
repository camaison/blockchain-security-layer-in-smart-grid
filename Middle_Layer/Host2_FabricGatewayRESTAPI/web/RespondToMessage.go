package web

import (
	"encoding/json"
	"fmt"
	"net/http"
    "time"
)

// RespondToMessage godoc
// @Summary Respond to a message on the ledger
// @Description Update or create a response in the ledger by its ID
// @Tags responses
// @Accept json
// @Produce json
// @Param id formData string true "ID of the Response"
// @Param subscribedContent formData string true "Subscribed Content as a JSON string"
// @Param publishedContent formData string true "Published Content as a JSON string"
// @Success 200 {string} string "Response Updated"
// @Failure 400 {string} string "Error occurred"
// @Router /respond [post]

func (setup *OrgSetup) RespondToMessage(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received RespondToMessage request")

    // Start timing
    start := time.Now()

    // Define a structure to match expected JSON payload
    type Request struct {
        ID                string                 `json:"id"`
        SubscribedContent map[string]interface{} `json:"subscribedContent"`
        PublishedContent  map[string]interface{} `json:"publishedContent"`
    }

    var requestData Request
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Convert the contents to JSON strings for chaincode interaction
    subscribedContentBytes, err := json.Marshal(requestData.SubscribedContent)
    if err != nil {
        http.Error(w, "JSON Marshal error for subscribedContent: "+err.Error(), http.StatusBadRequest)
        return
    }

    publishedContentBytes, err := json.Marshal(requestData.PublishedContent)
    if err != nil {
        http.Error(w, "JSON Marshal error for publishedContent: "+err.Error(), http.StatusBadRequest)
        return
    }

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    // Submit transaction to the ledger
    result, err := contract.SubmitTransaction("RespondToMessage", requestData.ID, string(subscribedContentBytes), string(publishedContentBytes))
    if err != nil {
        http.Error(w, "Error invoking RespondToMessage: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Calculate time spent
    end := time.Now()
    timeSpent := end.Sub(start).Seconds()

    // Broadcast the time spent
    broadcastTimeSpent(timeSpent)

    fmt.Fprintf(w, "Result: %s", result)
}

