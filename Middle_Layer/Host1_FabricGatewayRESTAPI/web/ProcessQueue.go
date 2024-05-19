package web

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
    "log"
)

// Define the types of jobs
type JobType int

const (
    UpdateMessageJob JobType = iota
    RespondToMessageJob
    ValidateMessageJob
)

// Job represents a queued job with its type and payload
type Job struct {
    Type    JobType
    Payload map[string]interface{}
}

var jobQueue []Job
var queueMutex sync.Mutex

// Enqueue a new job
func enqueueJob(job Job) {
    queueMutex.Lock()
    jobQueue = append(jobQueue, job)
    queueMutex.Unlock()
}

// Dequeue a job
func dequeueJob() (Job, bool) {
    queueMutex.Lock()
    defer queueMutex.Unlock()
    if len(jobQueue) == 0 {
        return Job{}, false
    }
    job := jobQueue[0]
    jobQueue = jobQueue[1:]
    return job, true
}

func processQueue(setup OrgSetup) {
    for {
        job, ok := dequeueJob()
        if !ok {
            time.Sleep(1 * time.Second) // Sleep if the queue is empty
            continue
        }

        switch job.Type {
        case UpdateMessageJob:
            setup.processUpdateMessage(job.Payload)
        case RespondToMessageJob:
            setup.processRespondToMessage(job.Payload)
        case ValidateMessageJob:
            setup.processValidateMessage(job.Payload)
        }
    }
}

// Placeholder for your actual processing functions
func (setup *OrgSetup) processUpdateMessage(data map[string]interface{}) {
    fmt.Printf("Processing UpdateMessage: %v\n", data)
    // Call the actual UpdateMessage function here
    id := data["id"].(string)
    messageType := data["messageType"].(string)
    messageContent := data["messageContent"].(map[string]interface{})

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    messageContentBytes, err := json.Marshal(messageContent)
    if err != nil {
        log.Println("JSON Marshal error:", err)
        return
    }

    _, err = contract.SubmitTransaction("UpdateMessage", id, string(messageContentBytes), messageType)
    if err != nil {
        log.Println("Error invoking UpdateMessage:", err)
        return
    }

    // Broadcast the updated data to all WebSocket clients
    allData, err := setup.getAllData()
    if err != nil {
        log.Println("Error fetching all data:", err)
        return
    }

    broadcastUpdate(allData)
}

func (setup *OrgSetup) processRespondToMessage(data map[string]interface{}) {
    fmt.Printf("Processing RespondToMessage: %v\n", data)
    // Call the actual RespondToMessage function here
    id := data["id"].(string)
    subscribedContent := data["subscribedContent"].(map[string]interface{})
    publishedContent := data["publishedContent"].(map[string]interface{})

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    subscribedContentBytes, err := json.Marshal(subscribedContent)
    if err != nil {
        log.Println("JSON Marshal error:", err)
        return
    }

    publishedContentBytes, err := json.Marshal(publishedContent)
    if err != nil {
        log.Println("JSON Marshal error:", err)
        return
    }

    _, err = contract.SubmitTransaction("RespondToMessage", id, string(subscribedContentBytes), string(publishedContentBytes))
    if err != nil {
        log.Println("Error invoking RespondToMessage:", err)
        return
    }

    // Broadcast the updated data to all WebSocket clients
    allData, err := setup.getAllData()
    if err != nil {
        log.Println("Error fetching all data:", err)
        return
    }

    broadcastUpdate(allData)
}

func (setup *OrgSetup) processValidateMessage(data map[string]interface{}) {
    fmt.Printf("Processing ValidateMessage: %v\n", data)
    // Call the actual ValidateMessage function here
    messageID := data["messageID"].(string)
    subscribedContent := data["subscribedContent"].(map[string]interface{})

    network := setup.Gateway.GetNetwork(setup.Channel)
    contract := network.GetContract(setup.Chaincode)

    subscribedContentBytes, err := json.Marshal(subscribedContent)
    if err != nil {
        log.Println("JSON Marshal error:", err)
        return
    }

    _, err = contract.EvaluateTransaction("ValidateMessage", messageID, string(subscribedContentBytes))
    if err != nil {
        log.Println("Error invoking ValidateMessage:", err)
        return
    }

    // Broadcast the updated data to all WebSocket clients
    allData, err := setup.getAllData()
    if err != nil {
        log.Println("Error fetching all data:", err)
        return
    }

    broadcastUpdate(allData)
}

func (setup *OrgSetup) enqueueValidateMessageHandler(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    job := Job{
        Type:    ValidateMessageJob,
        Payload: requestData,
    }

    enqueueJob(job)
    fmt.Fprintf(w, "ValidateMessage request queued")
}

func (setup *OrgSetup) enqueueRespondToMessageHandler(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    job := Job{
        Type:    RespondToMessageJob,
        Payload: requestData,
    }

    enqueueJob(job)
    fmt.Fprintf(w, "RespondToMessage request queued")
}

func (setup *OrgSetup) enqueueUpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "JSON Decode error: "+err.Error(), http.StatusBadRequest)
        return
    }

    job := Job{
        Type:    UpdateMessageJob,
        Payload: requestData,
    }

    enqueueJob(job)
    fmt.Fprintf(w, "UpdateMessage request queued")
}
