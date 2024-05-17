package web

import (
    "fmt"
    "log"
    "net/http"

    "github.com/hyperledger/fabric-gateway/pkg/client"
    "github.com/swaggo/http-swagger" // http-swagger middleware
    _ "rest-api-go/docs"             // This imports the generated Swagger docs
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

// Serve initializes and starts the HTTP server
func Serve(setups OrgSetup) {
    mux := http.NewServeMux()

    // Define routes for direct endpoints
    mux.HandleFunc("/update", setups.UpdateMessage)
    mux.HandleFunc("/respond", setups.RespondToMessage)
    mux.HandleFunc("/getAll", setups.GetAllData)
    mux.HandleFunc("/read", setups.ReadData)
    mux.HandleFunc("/validate", setups.ValidateMessage)
    mux.HandleFunc("/history", setups.GetTxnHistory)

    // Define routes for enqueue endpoints
    mux.HandleFunc("/enqueue/update", setups.enqueueUpdateMessageHandler)
    mux.HandleFunc("/enqueue/respond", setups.enqueueRespondToMessageHandler)
    mux.HandleFunc("/enqueue/validate", setups.enqueueValidateMessageHandler)

    // Serve Swagger documentation
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

    // Wrap the mux with the logging middleware
    loggedMux := loggingMiddleware(mux)

    // Start the queue processor in a separate goroutine
    go processQueue(setups)

    fmt.Println("Listening on http://localhost:3030/ ...")
    if err := http.ListenAndServe(":3030", loggedMux); err != nil {
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
