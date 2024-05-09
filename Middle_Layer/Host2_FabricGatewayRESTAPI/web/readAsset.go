package web

import (
	"fmt"
	"net/http"
)

// ReadAsset handles chaincode readAsset requests.
func (setup OrgSetup) ReadAsset(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received Read Asset request")
	queryParams := r.URL.Query()
	chainCodeName := queryParams.Get("chaincodeid")
	channelID := queryParams.Get("channelid")
	function := queryParams.Get("function")
	args := r.URL.Query()["args"]
	fmt.Printf("channel: %s, chaincode: %s, function: %s, args: %s\n", channelID, chainCodeName, function, args)
	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)
	evaluateResponse, err := contract.EvaluateTransaction(function, args...)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	fmt.Fprintf(w, "Response: %s", evaluateResponse)
}

//GET
//http://localhost:3000/readAsset?channelid=mychannel&chaincodeid=asset-transfer-basic&function=ReadAsset&args=asset1