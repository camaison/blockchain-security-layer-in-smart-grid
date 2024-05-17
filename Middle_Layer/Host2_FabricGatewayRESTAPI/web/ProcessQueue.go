package web

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/http/httptest"
    "sync"
    "time"
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
    defer queueMutex.Unlock()
    jobQueue = append(jobQueue, job)
    logQueue("Job added")
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
    logQueue("Job removed")
    return job, true
}

func logQueue(action string) {
    log.Printf("%s: Current Queue Length: %d\n", action, len(jobQueue))
    for i, job := range jobQueue {
        log.Printf("Queue[%d]: Type: %d, Payload: %v\n", i, job.Type, job.Payload)
    }
}

func processQueue(setup OrgSetup) {
    for {
        job, ok := dequeueJob()
        if !ok {
            time.Sleep(1 * time.Second) // Sleep if the queue is empty
            continue
        }

        log.Printf("Processing job: Type: %d, Payload: %v\n", job.Type, job.Payload)
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

// Process functions to call the actual handlers
func (setup *OrgSetup) processUpdateMessage(data map[string]interface{}) {
    // Convert the data map to JSON bytes
    requestData, err := json.Marshal(data)
    if err != nil {
        log.Printf("JSON Marshal error: %s\n", err)
        return
    }

    // Create a new HTTP request
    req, err := http.NewRequest("POST", "/update", bytes.NewReader(requestData))
    if err != nil {
        log.Printf("Error creating request: %s\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a ResponseRecorder to capture the response
    w := httptest.NewRecorder()
    setup.UpdateMessage(w, req)

    // Check the response
    res := w.Result()
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        log.Printf("Error processing update message: %s\n", res.Status)
    } else {
        log.Printf("UpdateMessage processed successfully: %s\n", res.Status)
    }
}

func (setup *OrgSetup) processRespondToMessage(data map[string]interface{}) {
    // Convert the data map to JSON bytes
    requestData, err := json.Marshal(data)
    if err != nil {
        log.Printf("JSON Marshal error: %s\n", err)
        return
    }

    // Create a new HTTP request
    req, err := http.NewRequest("POST", "/respond", bytes.NewReader(requestData))
    if err != nil {
        log.Printf("Error creating request: %s\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a ResponseRecorder to capture the response
    w := httptest.NewRecorder()
    setup.RespondToMessage(w, req)

    // Check the response
    res := w.Result()
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        log.Printf("Error processing respond message: %s\n", res.Status)
    } else {
        log.Printf("RespondToMessage processed successfully: %s\n", res.Status)
    }
}

func (setup *OrgSetup) processValidateMessage(data map[string]interface{}) {
    // Convert the data map to JSON bytes
    requestData, err := json.Marshal(data)
    if err != nil {
        log.Printf("JSON Marshal error: %s\n", err)
        return
    }

    // Create a new HTTP request
    req, err := http.NewRequest("POST", "/validate", bytes.NewReader(requestData))
    if err != nil {
        log.Printf("Error creating request: %s\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    // Create a ResponseRecorder to capture the response
    w := httptest.NewRecorder()
    setup.ValidateMessage(w, req)

    // Check the response
    res := w.Result()
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        log.Printf("Error processing validate message: %s\n", res.Status)
    } else {
        log.Printf("ValidateMessage processed successfully: %s\n", res.Status)
    }
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
