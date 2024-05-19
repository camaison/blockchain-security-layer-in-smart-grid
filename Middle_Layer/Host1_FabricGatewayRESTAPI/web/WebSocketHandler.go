package web

import (
	"encoding/json"
	"net/http"
	"log"
	"fmt"

	"github.com/gorilla/websocket" 
)

// WebSocket handler for real-time updates
func (setup *OrgSetup) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    clientsMutex.Lock()
    clients[conn] = true
    clientsMutex.Unlock()

    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            clientsMutex.Lock()
            delete(clients, conn)
            clientsMutex.Unlock()
            break
        }
    }
}

func broadcastUpdate(data interface{}) {
    clientsMutex.Lock()
    defer clientsMutex.Unlock()

    responseData, err := json.Marshal(data)
    if err != nil {
        log.Println("JSON Marshal error:", err)
        return
    }

    for client := range clients {
        err := client.WriteMessage(websocket.TextMessage, responseData)
        if err != nil {
            log.Println("Write error:", err)
            client.Close()
            delete(clients, client)
        }
    }
}

func (setup *OrgSetup) getAllData() (map[string]interface{}, error) {
    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    // Evaluate transaction using the GetAllData function from chaincode
    result, err := contract.EvaluateTransaction("GetAllData")
    if err != nil {
        return nil, fmt.Errorf("Error querying GetAllData: %s", err)
    }

    // Prepare the response to return JSON data
    var data map[string]interface{}
    if err := json.Unmarshal(result, &data); err != nil {
        return nil, fmt.Errorf("Error unmarshaling JSON data: %s", err)
    }

    return data, nil
}
