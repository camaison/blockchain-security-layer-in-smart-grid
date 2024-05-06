package web

import (
	"fmt"
	"net/http"
)

// DeleteAsset handles chaincode deleteAsset requests.
func (setup *OrgSetup) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received Delete Asset request")
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %s", err)
		return
	}
	chainCodeName := r.FormValue("chaincodeid")
	channelID := r.FormValue("channelid")
	function := r.FormValue("function")
	id := r.FormValue("id") // Assuming the ID of the asset is provided as a form parameter named 'id'.

	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)

	// Submitting a transaction to delete the asset with the given ID.
	_, err := contract.SubmitTransaction(function, id)
	if err != nil {
		fmt.Fprintf(w, "Error deleting asset: %s", err)
		return
	}
	fmt.Fprintf(w, "Asset deleted successfully")
}
