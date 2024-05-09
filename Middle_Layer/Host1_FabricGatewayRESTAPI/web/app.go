package web

import (
	"fmt"
	"net/http"



	"github.com/swaggo/http-swagger" 
	_ "github.com/camaison/blockchain-security-layer-in-smart-grid/Middle_Layer/Host1_FabricGatewayRESTAPI/docs"
)

// OrgSetup contains organization's config to interact with the network.
type OrgSetup struct {
	OrgName      string
	MSPID        string
	CryptoPath   string
	CertPath     string
	KeyPath      string
	TLSCertPath  string
	PeerEndpoint string
	GatewayPeer  string
	Gateway      client.Gateway
	Chaincode    string
	Channel      string
}


// @title Hyperledger Fabric Chaincode API
// @description This is a sample server for Hyperledger Fabric chaincode interaction.
// @version 1.0
// @host localhost:3000
// @BasePath /
// Serve starts http web server.
func Serve(setups OrgSetup) {
	http.HandleFunc("/update", setups.UpdateMessage)
	http.HandleFunc("/respond", setups.RespondToMessage)
	http.HandleFunc("/getAll", setups.GetAllData)
	http.HandleFunc("/read", setups.ReadData)
	http.HandleFunc("/validate", setups.ValidateMessage)
	http.HandleFunc("/history", setups.GetTxnHistory)
	
	// Serve Swagger
	url := httpSwagger.URL("http://localhost:3000/swagger/doc.json") // The url pointing to API definition
	http.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(url), // The url pointing to API definition
	))

	fmt.Println("Listening (http://localhost:3000/)...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Println(err)
	}
}
