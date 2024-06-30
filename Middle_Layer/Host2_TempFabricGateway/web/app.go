package web

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hyperledger/fabric-gateway/pkg/client"
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

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

// Serve initializes and starts the HTTP server
func Serve(setups OrgSetup) {
	mux := http.NewServeMux()

	// Define routes for direct endpoints
	mux.HandleFunc("/bookKeeping", setups.BookKeeping)
	mux.HandleFunc("/updateId", setups.UpdateIDs)
	mux.HandleFunc("/getState", setups.GetState)
	mux.HandleFunc("/read", setups.Read)
	mux.HandleFunc("/history", setups.GetHistory)
	mux.HandleFunc("/validate", setups.ValidateID)

	// WebSocket endpoint for real-time updates
	mux.HandleFunc("/ws", setups.handleWebSocket)

	// Serve the static HTML file
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// Wrap the mux with the logging middleware
	loggedMux := loggingMiddleware(mux)

	fmt.Println("Listening on http://localhost:3001/ ...")
	if err := http.ListenAndServe(":3001", loggedMux); err != nil {
		log.Fatal("ListenAndServe Error:", err)
	}
}

// loggingMiddleware logs all incoming HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
