#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <string.h>
#include <unistd.h>
#include <sys/time.h>
#include <time.h>
#include <curl/curl.h>

#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "goose_publisher.h"
#include "hal_thread.h"
#include "mms_value.h"
#include "linked_list.h"
#include "logging.h"  // Custom logging library

volatile int running = 1;
volatile int ipp_status = 0; // Global variable for the IPP status
volatile int update = 0; // Global variable for the IPP status
static uint32_t stNum = 0;
static uint32_t sqNum = 0;
static char subscribed_timestamp_str[64]; // Global variable for the timestamp string of the subscribed message
static char published_timestamp_str[64]; // Global variable for the timestamp string of the published message
static uint32_t subscribed_stNum = 0;
static char subscribed_data[1024] = "FALSE";

// Signal handler for graceful termination
static void sigint_handler(int signalId) {
    running = 0;
}

// Helper function to get formatted UTC timestamp
static void get_utc_timestamp(char *buffer, size_t buffer_len)
{
    struct timeval tv;
    gettimeofday(&tv, NULL);
    struct tm *tm_utc = gmtime(&tv.tv_sec);
    int microsec = tv.tv_usec;

    // Format the timestamp in ISO 8601 with microseconds and 'Z' for UTC time zone
    snprintf(buffer, buffer_len, "%04d-%02d-%02dT%02d:%02d:%02d.%06dZ",
             tm_utc->tm_year + 1900, tm_utc->tm_mon + 1, tm_utc->tm_mday,
             tm_utc->tm_hour, tm_utc->tm_min, tm_utc->tm_sec, microsec);
}

// Function to send a POST request with JSON data
void send_post_request(const char *url, const char *json_data)
{
    CURL *curl;
    CURLcode res;

    // Initialize curl
    curl_global_init(CURL_GLOBAL_DEFAULT);
    curl = curl_easy_init();
    if (curl) {
        // Set the URL for the POST request
        curl_easy_setopt(curl, CURLOPT_URL, url);

        // Set the POST data
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);

        // Set the Content-Type header
        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

        // Perform the request, res will get the return code
        res = curl_easy_perform(curl);
        
        // Check for errors
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
        } else {
            printf("\nPOST request sent successfully.\n");
        }

        // Clean up
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

void gooseListener(GooseSubscriber subscriber, void *parameter) {
    uint8_t src[6], dst[6];
    GooseSubscriber_getSrcMac(subscriber, src);
    GooseSubscriber_getDstMac(subscriber, dst);
    char srcStr[18], dstStr[18];
    sprintf(srcStr, "%02x:%02x:%02x:%02x:%02x:%02x", src[0], src[1], src[2], src[3], src[4], src[5]);
    sprintf(dstStr, "%02x:%02x:%02x:%02x:%02x:%02x", dst[0], dst[1], dst[2], dst[3], dst[4], dst[5]);

    char formattedTime[100];
    formatUtcTime(formattedTime, sizeof(formattedTime), GooseSubscriber_getTimestamp(subscriber));

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
    //        189, // This should be dynamic, based on actual data size
    //        GooseSubscriber_getGoCbRef(subscriber),
    //        GooseSubscriber_getStNum(subscriber),
    //        GooseSubscriber_getDataSetValues(subscriber) ? "TRUE" : "FALSE");

    get_utc_timestamp(subscribed_timestamp_str, sizeof(subscribed_timestamp_str));
    if (subscribed_stNum != GooseSubscriber_getStNum(subscriber)){
        subscribed_stNum = GooseSubscriber_getStNum(subscriber);
        update = 1;
        stNum++;
        sqNum = 0;
    }
    MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
    MmsValue_printToBuffer(values, subscribed_data, sizeof(subscribed_data));
    // Update subscribed_data with a boolean value
    size_t length = strlen(subscribed_data);
    if (subscribed_data[length - 2] == '1') {
        strcpy(subscribed_data, "TRUE");
    } else {
        strcpy(subscribed_data, "FALSE");
    }
}

// GOOSE publisher function
void publish(GoosePublisher publisher) {
    LinkedList dataSetValues = LinkedList_create();
    bool statusBool = (ipp_status == 1);
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum++);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
        log_error("Error sending GOOSE message");
    } else {
        log_debug("GOOSE message sent successfully: Status %s", statusBool ? "True" : "False");
    }

    get_utc_timestamp(published_timestamp_str, sizeof(published_timestamp_str));

    if (update){
        // Prepare JSON data for the POST request
        char json_data[4024];
        snprintf(json_data, sizeof(json_data),
                "{\"id\": \"IPP_ValidationMessage\", \"subscribedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}, \"publishedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}}",
                subscribed_timestamp_str, subscribed_stNum, subscribed_data,
                published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE");

        // Send the POST request
        send_post_request("http://192.168.37.139:3010/enqueue/respond", json_data);

        update = 0;
    }
    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

int main(int argc, char **argv) {
    signal(SIGINT, sigint_handler);
    char *interface = (argc > 1) ? argv[1] : "ens33";
    char gocbRef[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";
    char datSet[100] = "simpleIOGenericIO/LLN0$AnalogValues";
    char goID[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";
    log_info("Using interface %s", interface);

    GooseReceiver receiver = GooseReceiver_create();
    GooseReceiver_setInterfaceId(receiver, interface);
    GooseSubscriber subscriber = GooseSubscriber_create(goID, NULL);
    uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};
    GooseSubscriber_setDstMac(subscriber, dstMac);
    GooseSubscriber_setAppId(subscriber, 1000);
    GooseSubscriber_setListener(subscriber, gooseListener, NULL);
    GooseReceiver_addSubscriber(receiver, subscriber);
    GooseReceiver_start(receiver);

    CommParameters gooseCommParameters = {0};
    gooseCommParameters.appId = 1000;
    memcpy(gooseCommParameters.dstAddress, dstMac, 6);
    GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, interface);
    GoosePublisher_setGoCbRef(publisher, gocbRef);
    GoosePublisher_setConfRev(publisher, 1);
    GoosePublisher_setDataSetRef(publisher, datSet);
    GoosePublisher_setTimeAllowedToLive(publisher, 5000);
    GoosePublisher_setGoID(publisher, goID);
    GoosePublisher_setSimulation(publisher, false);
    GoosePublisher_setNeedsCommission(publisher, false);
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum);
    GoosePublisher_setConfRev(publisher, 1);

    while (running) {
        publish(publisher);
        Thread_sleep(3000); // Sleep for a second
    }

    GoosePublisher_destroy(publisher);
    GooseReceiver_stop(receiver);
    GooseReceiver_destroy(receiver);
    log_info("Application terminated gracefully");

    return 0;
}
