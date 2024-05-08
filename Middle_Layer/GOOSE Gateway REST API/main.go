package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Message represents the structure of the payload received.
type Message struct {
	Published  string `json:"published"`
	Subscribed string `json:"subscribed"`
}

// queue to hold the messages
var messageQueue []map[string]string
var mutex sync.Mutex

func enqueueMessage(message map[string]string) {
	mutex.Lock()
	messageQueue = append(messageQueue, message)
	mutex.Unlock()
}

func showMessageQueue() {
	mutex.Lock()
	fmt.Println("Current Queue:")
	for _, msg := range messageQueue {
		fmt.Println(msg)
	}
	mutex.Unlock()
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Create map from the received message
	messageMap := map[string]string{
		"published":  msg.Published,
		"subscribed": msg.Subscribed,
	}
	// Enqueue the messagez
	enqueueMessage(messageMap)
	// Print the queue in the terminal
	showMessageQueue()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message received and queued")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/enque", messageHandler).Methods("POST")

	fmt.Println("Server is running on port 3030")
	log.Fatal(http.ListenAndServe(":3030", router))
}
