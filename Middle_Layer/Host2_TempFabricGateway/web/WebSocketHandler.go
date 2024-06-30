package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

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

func broadcastTimeSpent(timeSpent float64) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	data := map[string]interface{}{
		"timeSpent": timeSpent,
	}
	message, err := json.Marshal(data)
	if err != nil {
		log.Println("JSON Marshal error:", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func broadcastMetrics(metrics map[string]float64) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	message, err := json.Marshal(metrics)
	if err != nil {
		log.Println("JSON Marshal error:", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
