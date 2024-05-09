package web

import (
	"fmt"
	"net/http"

)

func (setup *OrgSetup) GetAllAssets(w http.ResponseWriter, r *http.Request) {
    network := setup.Gateway.GetNetwork("mychannel")
    contract := network.GetContract("mychaincode")

    result, err := contract.EvaluateTransaction("GetAllAssets")
    if err != nil {
        fmt.Fprintf(w, "Error retrieving assets: %s", err)
        return
    }
    fmt.Fprintf(w, "Assets: %s", result)
}

//GET
//http://localhost:3000/readAsset?channelid=mychannel&chaincodeid=asset-transfer-basic&function=GetAllAssets