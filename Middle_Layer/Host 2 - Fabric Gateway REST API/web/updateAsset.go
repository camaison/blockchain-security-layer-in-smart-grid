package web

import (
	"fmt"
	"net/http"
	"strconv"
)

// UpdateAsset handles chaincode updateAsset requests.
func (setup *OrgSetup) UpdateAsset(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Received Update Asset request")
    if err := r.ParseForm(); err != nil {
        fmt.Fprintf(w, "ParseForm() err: %s", err)
        return
    }
    chainCodeName := r.FormValue("chaincodeid")
    channelID := r.FormValue("channelid")
    function := r.FormValue("function")
    id := r.FormValue("id")
    color := r.FormValue("color")
    size, err := strconv.Atoi(r.FormValue("size"))
    if err != nil {
        fmt.Fprintf(w, "Error parsing size: %s", err)
        return
    }
    owner := r.FormValue("owner")
    appraisedValue, err := strconv.Atoi(r.FormValue("appraisedValue"))
    if err != nil {
        fmt.Fprintf(w, "Error parsing appraised value: %s", err)
        return
    }
	
    network := setup.Gateway.GetNetwork(channelID)
    contract := network.GetContract(chainCodeName)
    _, err = contract.SubmitTransaction(function, id, color, strconv.Itoa(size), owner, strconv.Itoa(appraisedValue))
    if err != nil {
        fmt.Fprintf(w, "Error updating asset: %s", err)
        return
    }
    fmt.Fprintf(w, "Asset updated successfully")
}





/*
Method: POST
URL: http://localhost:3000/invoke
Body: Select x-www-form-urlencoded and add the following key-value pairs:
chaincodeid: The ID of your chaincode.
channelid: Your channel name.
function: UpdateAsset
id: asset1
color: red
size: 6
owner: Jane
appraisedValue: 3500
*/