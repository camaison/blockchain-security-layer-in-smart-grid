#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <string.h>
#include <unistd.h>
#include <sys/time.h>
#include <time.h>
#include <curl/curl.h>
#include <pthread.h>
#include <json-c/json.h> // Include json-c header

#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "goose_publisher.h"
#include "hal_thread.h"
#include "mms_value.h"
#include "linked_list.h"
#include "logging.h"  // Custom logging library
#include "hal_time.h"

volatile int running = 1;
volatile int ipp_status = 0; // Global variable for the IPP status
volatile int update = 0; // Global variable to trigger status update
static uint32_t stNum = 0;
static uint32_t sqNum = 0;
static char subscribed_timestamp_str[64]; // Global variable for the timestamp string of the subscribed message
static char api_timestamp_str[64];
static char api_subscribed_data[1024];
static char published_timestamp_str[64]; // Global variable for the timestamp string of the published message
static uint32_t subscribed_stNum = 0;
static char subscribed_data[1024] = "FALSE";
static GoosePublisher global_publisher;
static uint32_t previous_subscribed_stNum = 0; //THe stnum before the status update to be validated. subscribed stnum is reverted to this if the action is invalid

pthread_mutex_t lock; // Mutex for protecting shared variables

// Signal handler for graceful termination
static void sigint_handler(int signalId) {
    running = 0;
}

void log_error_with_retry(const char *message, int retry_count) {
    fprintf(stderr, "%s Retry count: %d\n", message, retry_count);
}

// Function to send a POST request with JSON data
void api_validate(const char *url, const char *json_data) {
    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    curl_global_init(CURL_GLOBAL_DEFAULT);
    do {
        curl = curl_easy_init();
        if (curl) {
            curl_easy_setopt(curl, CURLOPT_URL, url);
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);

            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK) {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count); // Exponential backoff
            } else {
                retry_count = max_retries; // Exit the loop if successful
            }

            curl_slist_free_all(headers);
            curl_easy_cleanup(curl);
        }
    } while (res != CURLE_OK && retry_count < max_retries);

    curl_global_cleanup();
}

void formatUtcTime(char* buffer, size_t buffer_size, uint64_t epoch_ms) {
    time_t rawtime = epoch_ms / 1000;
    struct tm * ptm = gmtime(&rawtime);

    strftime(buffer, buffer_size, "%b %d, %Y %H:%M:%S", ptm);
    int milliseconds = epoch_ms % 1000;
    char ms_buffer[50];
    sprintf(ms_buffer, ".%03d UTC", milliseconds);
    strcat(buffer, ms_buffer);
}

// Function to send a non-blocking POST request with JSON data
void api_update(const char* timestamp, uint32_t stNum, const char* allData, const char* update_status) {
    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    curl_global_init(CURL_GLOBAL_ALL);
    do {
        curl = curl_easy_init();
        if(curl) {
            json_object *jobj = json_object_new_object();
            json_object *jmessageContent = json_object_new_object();

            json_object_object_add(jmessageContent, "t", json_object_new_string(timestamp));
            json_object_object_add(jmessageContent, "stNum", json_object_new_int(stNum));
            json_object_object_add(jmessageContent, "allData", json_object_new_string(allData));

            json_object_object_add(jobj, "id", json_object_new_string("IPP_PubMessage"));
            json_object_object_add(jobj, "messageType", json_object_new_string(update_status));
            json_object_object_add(jobj, "messageContent", jmessageContent);

            const char *json_data = json_object_to_json_string(jobj);
            curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.145:3000/update");
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
            curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, 5000L);

            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK) {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count);
            } else {
                retry_count = max_retries;
            }

            curl_slist_free_all(headers);
            curl_easy_cleanup(curl);
            json_object_put(jobj);
        }
    } while (res != CURLE_OK && retry_count < max_retries);

    curl_global_cleanup();
}

// Callback function to write response data
size_t write_callback(void *ptr, size_t size, size_t nmemb, void *stream) {
    size_t total_size = size * nmemb;
    if(total_size < 1024) {
        memcpy(stream, ptr, total_size);
        ((char*)stream)[total_size] = '\0'; // Null-terminate the string
    }
    return total_size;
}

void publish(GoosePublisher publisher);

void* handle_update(void* arg) {
    char json_data[4024];
    bool statusBool = (ipp_status == 1);

    Thread_sleep(5000); // Sleep for 5 seconds

    snprintf(json_data, sizeof(json_data),
             "{\"id\": \"IPP_ValidationMessage\", \"subscribedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}, \"publishedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}}",
             api_timestamp_str, subscribed_stNum, api_subscribed_data,
             published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE");

    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    curl_global_init(CURL_GLOBAL_DEFAULT);
    do {
        curl = curl_easy_init();
        if(curl) {
            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");

            curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.145:3000/respond");
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            char response_buffer[1024] = {0};
            curl_easy_setopt(curl, CURLOPT_WRITEDATA, response_buffer);
            curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK) {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count);
            } else {
                printf("%s\n", response_buffer);
                retry_count = max_retries;
            }

            curl_slist_free_all(headers);
            curl_easy_cleanup(curl);

            if (strcmp(response_buffer, "Result: Invalid") == 0) {
                pthread_mutex_lock(&lock);

                stNum++;
                sqNum = 0;

                ipp_status = !ipp_status;
                statusBool = (ipp_status == 1);

                subscribed_stNum = previous_subscribed_stNum;

                pthread_mutex_unlock(&lock);

                publish(global_publisher);

                api_update(published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE", "Corrective");
            }
        }
    } while (res != CURLE_OK && retry_count < max_retries);

    curl_global_cleanup();

    return NULL;
}

// Listener for GOOSE messages
void gooseListener(GooseSubscriber subscriber, void *parameter) {
    uint8_t src[6], dst[6];
    GooseSubscriber_getSrcMac(subscriber, src);
    GooseSubscriber_getDstMac(subscriber, dst);
    char srcStr[18], dstStr[18];
    sprintf(srcStr, "%02x:%02x:%02x:%02x:%02x:%02x", src[0], src[1], src[2], src[3], src[4], src[5]);
    sprintf(dstStr, "%02x:%02x:%02x:%02x:%02x:%02x", dst[0], dst[1], dst[2], dst[3], dst[4], dst[5]);

    formatUtcTime(subscribed_timestamp_str, sizeof(subscribed_timestamp_str), GooseSubscriber_getTimestamp(subscriber));

    if (subscribed_stNum < GooseSubscriber_getStNum(subscriber)){
        previous_subscribed_stNum = subscribed_stNum;
        subscribed_stNum = GooseSubscriber_getStNum(subscriber);
        update = 1;
        memcpy(api_timestamp_str, subscribed_timestamp_str, sizeof(subscribed_timestamp_str));
        stNum++;
        sqNum = 0;
        MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
        MmsValue_printToBuffer(values, subscribed_data, sizeof(subscribed_data));
        size_t length = strlen(subscribed_data);
        if (length == 6) {
            strcpy(subscribed_data, "TRUE");
            ipp_status = 0;
        } else {
            strcpy(subscribed_data, "FALSE");
            ipp_status = 1;
        }
        memcpy(api_subscribed_data, subscribed_data, sizeof(subscribed_data));

        printf("Previously Subscribed: %u\nCurrently Subscribed: %u\n", previous_subscribed_stNum, subscribed_stNum);
        printf("{\n"
           "\"t\": \"%s\",\n"
           "\"stNum\": %u,\n"
           "\"allData\": %s\n"
           "}\n",
           subscribed_timestamp_str,
           subscribed_stNum,
           subscribed_data);
    }
}

// GOOSE publisher function
void publish(GoosePublisher publisher) {
    LinkedList dataSetValues = LinkedList_create();
    bool statusBool = (ipp_status == 1);
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum++);

    uint64_t currentTime = Hal_getTimeInMs();
    GoosePublisher_setTimestamp(publisher, currentTime);

    formatUtcTime(published_timestamp_str, sizeof(published_timestamp_str), currentTime);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
        log_error("Error sending GOOSE message");
    }

    if (update) {
        global_publisher = publisher;
        update = 0;
        pthread_t update_thread;
        pthread_create(&update_thread, NULL, handle_update, NULL);
        pthread_detach(update_thread); // Automatically reclaim thread resources when done
    }

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

int main(int argc, char **argv) {
    signal(SIGINT, sigint_handler);
    pthread_mutex_init(&lock, NULL); // Initialize the mutex

    char *interface = (argc > 1) ? argv[1] : "ens33";
    char gocbRef[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";
    char datSet[100] = "simpleIOGenericIO/LLN0$AnalogValues";
    char goID[100] = "simpleIOGenericIO/LLN0$GO$gcbAnalogValues";
    
    log_info("Using interface %s", interface);

    GooseReceiver receiver = GooseReceiver_create();
    if (receiver == NULL) {
        log_error("Failed to create GooseReceiver");
        return EXIT_FAILURE;
    }

    GooseReceiver_setInterfaceId(receiver, interface);
    GooseSubscriber subscriber = GooseSubscriber_create(goID, NULL);
    if (subscriber == NULL) {
        log_error("Failed to create GooseSubscriber");
        GooseReceiver_destroy(receiver);
        return EXIT_FAILURE;
    }

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
    
    if (publisher == NULL) {
        log_error("Failed to create GoosePublisher");
        GooseReceiver_stop(receiver);
        GooseReceiver_destroy(receiver);
        GooseSubscriber_destroy(subscriber);
        return EXIT_FAILURE;
    }

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
        pthread_mutex_lock(&lock);
        publish(publisher);
        pthread_mutex_unlock(&lock);
        Thread_sleep(5000);
    }

    GoosePublisher_destroy(publisher);
    GooseReceiver_stop(receiver);
    GooseReceiver_destroy(receiver);
    pthread_mutex_destroy(&lock);
    log_info("Application terminated gracefully");

    return 0;
}
