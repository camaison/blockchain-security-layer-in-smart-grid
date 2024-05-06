#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <signal.h>
#include <time.h>
#include <curl/curl.h>
#include <json-c/json.h>


#include "mms_value.h"
#include "goose_publisher.h"
#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "hal_thread.h"
#include "logging.h"  // Ensure you have a logging library

static volatile int running = 1;
static int rdso_status = 1;
static uint32_t stNum = 0;
static uint32_t sqNum = 0;
 
char gocbRef[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";
char datSet[100] = "simpleIOGenericIO/LLN0$AnalogValues";
char goID[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";

// Signal handler for graceful termination
static void sigint_handler(int signalId) {
    running = 0;
}

void api_push(const char* published_message)
{
    CURL *curl;
    CURLcode res;

    curl_global_init(CURL_GLOBAL_ALL);
    curl = curl_easy_init();
    if(curl) {
        const char *published = published_message;
        const char *subscribed = "Sample Subscribed!";

        // JSON payload
        char jsonData[1000];
        snprintf(jsonData, sizeof(jsonData), "{\"published\":\"%s\", \"subscribed\":\"%s\"}", published, subscribed);

        curl_easy_setopt(curl, CURLOPT_URL, "http://localhost:3030/api/message");
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonData);

        // Set headers for JSON
        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        // Perform the request, res will get the return code
        printf("Is before error?\n\n");
        res = curl_easy_perform(curl);

        // Check for errors
        if(res != CURLE_OK)
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));

        // Cleanup
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
    }
    curl_global_cleanup();
}

void formatUtcTime(char* buffer, size_t buffer_size, uint64_t epoch_ms) {
    time_t rawtime = epoch_ms / 1000;
    struct tm * ptm = gmtime(&rawtime);

    strftime(buffer, buffer_size, "%b %d, %Y %H:%M:%S", ptm);
    // Add milliseconds manually since strftime doesn't support them
    int milliseconds = epoch_ms % 1000;
    char ms_buffer[50];
    sprintf(ms_buffer, ".%03d UTC", milliseconds);
    strcat(buffer, ms_buffer);
}

// Publish function
void publish(GoosePublisher publisher) {
    //re-add api push functionality when fixed
    LinkedList dataSetValues = LinkedList_create();
    bool statusBool = (rdso_status == 1); // True for CLOSED, False for OPEN
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

    if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
        log_error("Error sending GOOSE message");
    } else {
        // Create JSON object
        json_object *jobj = json_object_new_object();
        json_object *jgoID = json_object_new_string(goID);
        json_object *jtime = json_object_new_string("May  4, 2024 18:44:48.280999958 UTC"); // Should be dynamically set
        json_object *jsize = json_object_new_int(189); // Should be dynamically calculated
        json_object *jgocbRef = json_object_new_string(gocbRef);
        json_object *jstNum = json_object_new_int(stNum);
        json_object *jallData = json_object_new_string("TRUE"); // Change the value to be dynamic based on statusBool
        // Construct the JSON object
        json_object_object_add(jobj, "goID", jgoID);
        json_object_object_add(jobj, "t", jtime);
        json_object_object_add(jobj, "size", jsize);
        json_object_object_add(jobj, "gocbRef", jgocbRef);
        json_object_object_add(jobj, "stNum", jstNum);
        json_object_object_add(jobj, "allData", jallData);

        // Convert to JSON string
        const char *pubPayload = json_object_to_json_string(jobj);
        printf("Generated JSON: %s\n", pubPayload); // Debugging line


        // Use the JSON string to push to the API
        //api_push(pubPayload);

        // Clean up
        //json_object_put(jobj);
    }
    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

// void publish(GoosePublisher publisher) {
//     LinkedList dataSetValues = LinkedList_create();
//     bool statusBool = (rdso_status == 1); // True for CLOSED, False for OPEN
//     LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

//     if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
//         log_error("Error sending GOOSE message");
//     } else {
//         json_object *jobj = json_object_new_object();
//         json_object_object_add(jobj, "goID", json_object_new_string(goID));
//         json_object_object_add(jobj, "t", json_object_new_string("May  4, 2024 18:44:48.280999958 UTC"));
//         json_object_object_add(jobj, "size", json_object_new_int(189));
//         json_object_object_add(jobj, "gocbRef", json_object_new_string(gocbRef));
//         json_object_object_add(jobj, "stNum", json_object_new_int(stNum));
//         json_object_object_add(jobj, "allData", json_object_new_string("TRUE"));

//         const char *pubPayload = json_object_to_json_string(jobj);
//         printf("Generated JSON: %s\n", pubPayload); // Debugging line

//         api_push(pubPayload);

//         json_object_put(jobj);
//     }
//     LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
// }

// Listener function for receiving GOOSE messages
// static void gooseListener(GooseSubscriber subscriber, void *parameter) {
//     printf("GOOSE event:\n");
//     printf("  stNum: %u sqNum: %u\n", GooseSubscriber_getStNum(subscriber), GooseSubscriber_getSqNum(subscriber));
//     printf("  timeToLive: %u\n", GooseSubscriber_getTimeAllowedToLive(subscriber));

//     uint64_t timestamp = GooseSubscriber_getTimestamp(subscriber);
//     printf("  timestamp: %u.%u\n", (uint32_t)(timestamp / 1000), (uint32_t)(timestamp % 1000));
//     printf("  message is %s\n", GooseSubscriber_isValid(subscriber) ? "valid" : "INVALID");

//     MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
//     char buffer[1024];
//     MmsValue_printToBuffer(values, buffer, 1024);
//     printf("  allData: %s\n", buffer);
// }



// void gooseListener(GooseSubscriber subscriber, void *parameter) {
//     uint8_t src[6], dst[6];
//     GooseSubscriber_getSrcMac(subscriber, src);
//     GooseSubscriber_getDstMac(subscriber, dst);
//     char srcStr[18], dstStr[18];
//     sprintf(srcStr, "%02x:%02x:%02x:%02x:%02x:%02x", src[0], src[1], src[2], src[3], src[4], src[5]);
//     sprintf(dstStr, "%02x:%02x:%02x:%02x:%02x:%02x", dst[0], dst[1], dst[2], dst[3], dst[4], dst[5]);

//     char formattedTime[100];
//     formatUtcTime(formattedTime, sizeof(formattedTime), GooseSubscriber_getTimestamp(subscriber));

//     printf("{\n"
//            "\"goID\": \"%s\",\n"
//            "\"Src\": \"%s\",\n"
//            "\"Dst\": \"%s\",\n"
//            "\"t\": \"%s\",\n"
//            "\"size\": %d,\n"
//            "\"gocbRef\": \"%s\",\n"
//            "\"stNum\": %u,\n"
//            "\"allData\": %s\n"
//            "}\n",
//            GooseSubscriber_getGoId(subscriber),
//            srcStr, dstStr,
//            formattedTime,
//            189, // This should be dynamic, based on actual data size
//            GooseSubscriber_getGoCbRef(subscriber),
//            GooseSubscriber_getStNum(subscriber),
//            GooseSubscriber_getDataSetValues(subscriber) ? "TRUE" : "FALSE");
// }
void gooseListener(GooseSubscriber subscriber, void *parameter) {
    uint8_t src[6], dst[6];
    GooseSubscriber_getSrcMac(subscriber, src);
    GooseSubscriber_getDstMac(subscriber, dst);
    char srcStr[18], dstStr[18];
    sprintf(srcStr, "%02x:%02x:%02x:%02x:%02x:%02x", src[0], src[1], src[2], src[3], src[4], src[5]);
    sprintf(dstStr, "%02x:%02x:%02x:%02x:%02x:%02x", dst[0], dst[1], dst[2], dst[3], dst[4], dst[5]);

    char formattedTime[100];
    formatUtcTime(formattedTime, sizeof(formattedTime), GooseSubscriber_getTimestamp(subscriber));


    // MmsValue *value = GooseSubscriber_getDataSetValues(subscriber);
    // bool allData = false;
    // if (value && MmsValue_getType(value) == MMS_BOOLEAN) {
    //     allData = MmsValue_getBoolean(value);
    // }

    // printf("{\n"
    //        "\"goID\": \"%s\",\n"
    //        "\"Src\": \"%s\",\n"
    //        "\"Dst\": \"%s\",\n"
    //        "\"t\": \"%s\",\n"
    //        "\"size\": %d,\n"
    //        "\"gocbRef\": \"%s\",\n"
    //        "\"stNum\": %u,\n"
    //        "\"allData\": %s\n"
    //        "}\n",
    //        GooseSubscriber_getGoId(subscriber),
    //        srcStr, dstStr,
    //        formattedTime,
    //        189, // Dynamically obtained packet size
    //        GooseSubscriber_getGoCbRef(subscriber),
    //        GooseSubscriber_getStNum(subscriber),
    //        allData ? "TRUE" : "FALSE");
}


int main(int argc, char **argv) {
   
    char *interface = (argc > 1) ? argv[1] : "ens33";
    log_info("Using interface %s", interface);

    signal(SIGINT, sigint_handler);

    CommParameters gooseCommParameters = {0};
    gooseCommParameters.appId = 1000;
    gooseCommParameters.dstAddress[0] = 0x01;
    gooseCommParameters.dstAddress[1] = 0x0c;
    gooseCommParameters.dstAddress[2] = 0xcd;
    gooseCommParameters.dstAddress[3] = 0x01;
    gooseCommParameters.dstAddress[4] = 0x00;
    gooseCommParameters.dstAddress[5] = 0x01;

    GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, interface);
    GoosePublisher_setGoCbRef(publisher, gocbRef);
    GoosePublisher_setConfRev(publisher, 1);
    GoosePublisher_setDataSetRef(publisher, datSet);
    GoosePublisher_setTimeAllowedToLive(publisher, 500);

    GooseReceiver receiver = GooseReceiver_create();
    GooseReceiver_setInterfaceId(receiver, interface);
    GooseSubscriber subscriber = GooseSubscriber_create(goID, NULL);
    uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};
    GooseSubscriber_setDstMac(subscriber, dstMac);
    GooseSubscriber_setAppId(subscriber, 1000);
    GooseSubscriber_setListener(subscriber, gooseListener, NULL);
    GooseReceiver_addSubscriber(receiver, subscriber);
    GooseReceiver_start(receiver);

    int toggleCounter = 0;
    while (running) {
        if (GooseReceiver_isRunning(receiver)) {
            if (toggleCounter++ % 10 == 0) {
                int old_status = rdso_status;
                rdso_status = !rdso_status;
                if (old_status != rdso_status) {
                    stNum++;
                    sqNum = 0;
                }
            } else {
                sqNum++;
            }
            GoosePublisher_setStNum(publisher, stNum);
            GoosePublisher_setSqNum(publisher, sqNum);

            publish(publisher);
            Thread_sleep(1000); // Sleep for a second
        }
    }

    GoosePublisher_destroy(publisher);
    GooseReceiver_stop(receiver);
    GooseReceiver_destroy(receiver);
    log_info("Application terminated gracefully");

    return 0;
}
