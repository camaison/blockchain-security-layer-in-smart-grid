package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %s", err)
		return
	}

	id := r.FormValue("id")

	// Parse the subscribedContent from JSON string in the form data
	subscribedContent := make(map[string]interface{})
	if err := json.Unmarshal([]byte(r.FormValue("subscribedContent")), &subscribedContent); err != nil {
		fmt.Fprintf(w, "JSON Unmarshal error for subscribedContent: %s", err)
		return
	}

	// Parse the publishedContent from JSON string in the form data
	publishedContent := make(map[string]interface{})
	if err := json.Unmarshal([]byte(r.FormValue("publishedContent")), &publishedContent); err != nil {
		fmt.Fprintf(w, "JSON Unmarshal error for publishedContent: %s", err)
		return
	}

	network := setup.Gateway.GetNetwork(setup.Channel)
	contract := network.GetContract(setup.Chaincode)

	// Prepare subscribedContent and publishedContent for chaincode interaction
	subscribedContentBytes, err := json.Marshal(subscribedContent)
	if err != nil {
		fmt.Fprintf(w, "JSON Marshal error for subscribedContent: %s", err)
		return
	}
	publishedContentBytes, err := json.Marshal(publishedContent)
	if err != nil {
		fmt.Fprintf(w, "JSON Marshal error for publishedContent: %s", err)
		return
	}

	// Submit transaction to the ledger
	result, err := contract.SubmitTransaction("RespondToMessage", id, string(subscribedContentBytes), string(publishedContentBytes))
	if err != nil {
		fmt.Fprintf(w, "Error invoking RespondToMessage: %s", err)
		return
	}
	fmt.Fprintf(w, "Response Updated: Status: %s", result)
}
