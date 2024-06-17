#define _POSIX_C_SOURCE 199309L // Add this line at the top

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
#include "logging.h" // Custom logging library
#include "hal_time.h"

volatile int running = 1;
volatile int ipp_status = 0; // Global variable for the IPP status
volatile int update = 0;     // Global variable to trigger status update
static uint32_t stNum = 0;
static uint32_t sqNum = 0;
static char subscribed_timestamp_str[64]; // Global variable for the timestamp string of the subscribed message
static char api_timestamp_str[64];
static char api_subscribed_data[1024];
static char published_timestamp_str[64]; // Global variable for the timestamp string of the published message
static uint32_t subscribed_stNum = 0;
static char subscribed_data[1024] = "FALSE";
static GoosePublisher global_publisher;
static uint32_t previous_subscribed_stNum = 0; // The stnum before the status update to be validated. subscribed stnum is reverted to this if the action is invalid
static char subscribed_goID[100];              // Global variable for the timestamp string of the published message

char gocbRef[100] = "IPP/LLN0$GO$gcbAnalogValues";
char datSet[100] = "IPP/LLN0$AnalogValues";
char goID[100] = "IPP";
char goIDListenerRDSO[100] = "RDSO/LLN0$GO$gcbAnalogValues";
char goIDListenerX[100] = "X/LLN0$GO$gcbAnalogValues";

pthread_mutex_t lock; // Mutex for protecting shared variables

// Signal handler for graceful termination
static void sigint_handler(int signalId)
{
    running = 0;
}

void log_error_with_retry(const char *message, int retry_count)
{
    fprintf(stderr, "%s Retry count: %d\n", message, retry_count);
}

// Function to send a POST request with JSON data
void api_validate(const char *url, const char *json_data)
{
    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    curl_global_init(CURL_GLOBAL_DEFAULT);
    do
    {
        curl = curl_easy_init();
        if (curl)
        {
            curl_easy_setopt(curl, CURLOPT_URL, url);
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);

            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK)
            {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count); // Exponential backoff
            }
            else
            {
                retry_count = max_retries; // Exit the loop if successful
            }

            curl_slist_free_all(headers);
            curl_easy_cleanup(curl);
        }
    } while (res != CURLE_OK && retry_count < max_retries);

    curl_global_cleanup();
}

void formatUtcTime(char *buffer, size_t buffer_size, uint64_t epoch_ms)
{
    time_t rawtime = epoch_ms / 1000;
    struct tm *ptm = gmtime(&rawtime);

    strftime(buffer, buffer_size, "%b %d, %Y %H:%M:%S", ptm);
    int milliseconds = epoch_ms % 1000;
    char ms_buffer[50];
    sprintf(ms_buffer, ".%03d UTC", milliseconds);
    strcat(buffer, ms_buffer);
}

// Function to send a non-blocking POST request with JSON data
void api_update(const char *timestamp, uint32_t stNum, const char *allData, const char *update_status)
{
    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    curl_global_init(CURL_GLOBAL_ALL);
    do
    {
        curl = curl_easy_init();
        if (curl)
        {
            json_object *jobj = json_object_new_object();
            json_object *jmessageContent = json_object_new_object();

            json_object_object_add(jmessageContent, "t", json_object_new_string(timestamp));
            json_object_object_add(jmessageContent, "stNum", json_object_new_int(stNum));
            json_object_object_add(jmessageContent, "allData", json_object_new_string(allData));

            json_object_object_add(jobj, "id", json_object_new_string("IPP_PubMessage"));
            json_object_object_add(jobj, "messageType", json_object_new_string(update_status));
            json_object_object_add(jobj, "messageContent", jmessageContent);

            const char *json_data = json_object_to_json_string(jobj);
            curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.139:3001/update");
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
            curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, 5000L);

            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK)
            {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count);
            }
            else
            {
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
size_t write_callback(void *ptr, size_t size, size_t nmemb, void *stream)
{
    size_t total_size = size * nmemb;
    if (total_size < 1024)
    {
        memcpy(stream, ptr, total_size);
        ((char *)stream)[total_size] = '\0'; // Null-terminate the string
    }
    return total_size;
}

void publish(GoosePublisher publisher);

void *handle_update(void *arg)
{
    char json_data[4024];
    char query_data[4024];
    bool statusBool = (ipp_status == 1);

    struct timespec start, end;

    // Thread_sleep(3000); // Sleep for 3 seconds

    snprintf(query_data, sizeof(query_data),
             "{\"messageID\": \"RDSO_PubMessage\", \"subscribedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}}",
             api_timestamp_str, subscribed_stNum, api_subscribed_data);

    snprintf(json_data, sizeof(json_data),
             "{\"id\": \"IPP_ValidationMessage\", \"subscribedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}, \"publishedContent\": {\"t\": \"%s\", \"stNum\": %u, \"allData\": \"%s\"}}",
             api_timestamp_str, subscribed_stNum, api_subscribed_data,
             published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE");

    json_object *jobj = json_object_new_object();
    json_object_object_add(jobj, "value", json_object_new_string(subscribed_goID));
    const char *json_data_id = json_object_to_json_string(jobj);

    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;

    // Record start time
    clock_gettime(CLOCK_REALTIME, &start);
    curl_global_init(CURL_GLOBAL_DEFAULT);

    curl = curl_easy_init();
    if (curl)
    {
        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Content-Type: application/json");

        // curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.139:3000/validate");
        // curl_easy_setopt(curl, CURLOPT_POSTFIELDS, query_data);
        curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.145:3001/idValidate");
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data_id);
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

        char response_buffer[1024] = {0};
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, response_buffer);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);

        res = curl_easy_perform(curl);
        if (res != CURLE_OK)
        {
            log_error_with_retry("curl_easy_perform() failed", retry_count);
            retry_count++;
            Thread_sleep(2000 * retry_count);
        }
        else
        {
            printf("API RESPONSE:  %s", response_buffer);
            retry_count = max_retries;
        }

        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);

        // Record end time
        clock_gettime(CLOCK_REALTIME, &end);

        // Calculate time difference
        double time_spent = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
        printf("Time taken for Validation: %.9f seconds\n\n", time_spent);
        /*response:
        {
            "isValid": true
        } */

        json_object *jobj = json_tokener_parse(response_buffer);
        json_object *jisValid = NULL;
        json_object_object_get_ex(jobj, "isValid", &jisValid);
        bool isValid = json_object_get_boolean(jisValid);
        json_object_put(jobj);

        if (!isValid)
        {
            pthread_mutex_lock(&lock);

            stNum++;
            sqNum = 0;

            ipp_status = !ipp_status;
            statusBool = (ipp_status == 1);

            subscribed_stNum = previous_subscribed_stNum;

            pthread_mutex_unlock(&lock);

            publish(global_publisher);
        }
    }
    while (res != CURLE_OK && retry_count < max_retries)
        ;

    // Book Keeping Part of the Thread!!!
    clock_gettime(CLOCK_REALTIME, &start);
    curl_global_init(CURL_GLOBAL_DEFAULT);
    do
    {
        curl = curl_easy_init();
        if (curl)
        {
            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Content-Type: application/json");

            curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.139:3001/respond");
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

            char response_buffer[1024] = {0};
            curl_easy_setopt(curl, CURLOPT_WRITEDATA, response_buffer);
            curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK)
            {
                log_error_with_retry("curl_easy_perform() failed", retry_count);
                retry_count++;
                Thread_sleep(2000 * retry_count);
            }
            else
            {
                // printf("%s\n", response_buffer);
                retry_count = max_retries;
            }

            curl_slist_free_all(headers);
            curl_easy_cleanup(curl);

            if (strcmp(response_buffer, "Result: Invalid") == 0)
            {
                api_update(published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE", "Corrective");
            }

            // Record end time
            clock_gettime(CLOCK_REALTIME, &end);

            // Calculate time difference
            // double time_spent = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
            // printf("Time taken for Bookkeeping Process: %.9f seconds\n", time_spent);
        }
    } while (res != CURLE_OK && retry_count < max_retries);
    curl_global_cleanup();

    return NULL;
}

// Listener for GOOSE messages
void gooseListener(GooseSubscriber subscriber, void *parameter)
{
    uint8_t src[6], dst[6];
    GooseSubscriber_getSrcMac(subscriber, src);
    GooseSubscriber_getDstMac(subscriber, dst);

    formatUtcTime(subscribed_timestamp_str, sizeof(subscribed_timestamp_str), GooseSubscriber_getTimestamp(subscriber));

    if (subscribed_stNum < GooseSubscriber_getStNum(subscriber))
    {
        memcpy(subscribed_goID, GooseSubscriber_getGoId(subscriber), 100);
        previous_subscribed_stNum = subscribed_stNum;
        subscribed_stNum = GooseSubscriber_getStNum(subscriber);
        update = 1;
        memcpy(api_timestamp_str, subscribed_timestamp_str, sizeof(subscribed_timestamp_str));
        stNum++;
        sqNum = 0;
        MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
        MmsValue_printToBuffer(values, subscribed_data, sizeof(subscribed_data));
        size_t length = strlen(subscribed_data);
        if (length == 6)
        {
            strcpy(subscribed_data, "TRUE");
            ipp_status = 0;
        }
        else
        {
            strcpy(subscribed_data, "FALSE");
            ipp_status = 1;
        }
        memcpy(api_subscribed_data, subscribed_data, sizeof(subscribed_data));

        // printf("Previously Subscribed: %u\nCurrently Subscribed: %u\n", previous_subscribed_stNum, subscribed_stNum);
        printf("\n***MESSAGE SUBSCRIBED***\n{\n"
               "\"goID\": \"%s\",\n"
               "\"allData\": %s\n"
               "}\n",
               subscribed_goID,
               subscribed_data);
    }
}

// GOOSE publisher function
void publish(GoosePublisher publisher)
{
    LinkedList dataSetValues = LinkedList_create();
    bool statusBool = (ipp_status == 1);
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum++);

    uint64_t currentTime = Hal_getTimeInMs();
    GoosePublisher_setTimestamp(publisher, currentTime);

    formatUtcTime(published_timestamp_str, sizeof(published_timestamp_str), currentTime);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1)
    {
        log_error("Error sending GOOSE message");
    }

    if (update)
    {
        global_publisher = publisher;
        update = 0;
        pthread_t update_thread;
        pthread_create(&update_thread, NULL, handle_update, NULL);
        pthread_detach(update_thread); // Automatically reclaim thread resources when done
    }

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

int main(int argc, char **argv)
{
    signal(SIGINT, sigint_handler);
    pthread_mutex_init(&lock, NULL); // Initialize the mutex

    char *interface = (argc > 1) ? argv[1] : "ens33";

    log_info("Using interface %s", interface);

    GooseReceiver receiver = GooseReceiver_create();
    if (receiver == NULL)
    {
        log_error("Failed to create GooseReceiver");
        return EXIT_FAILURE;
    }

    GooseReceiver_setInterfaceId(receiver, interface);
    GooseSubscriber subscriberRDSO = GooseSubscriber_create(goIDListenerRDSO, NULL);
    GooseSubscriber subscriberX = GooseSubscriber_create(goIDListenerX, NULL);
    GooseSubscriber subscriber = GooseSubscriber_create(goID, NULL);

    uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};

    if (subscriberRDSO == NULL)
    {
        log_error("Failed to create GooseSubscriber RDSO");
        GooseReceiver_destroy(receiver);
        return EXIT_FAILURE;
    }
    else
    {
        GooseSubscriber_setDstMac(subscriberRDSO, dstMac);
        GooseSubscriber_setAppId(subscriberRDSO, 1000);
        GooseSubscriber_setListener(subscriberRDSO, gooseListener, NULL);
        GooseReceiver_addSubscriber(receiver, subscriberRDSO);
    }

    if (subscriberX == NULL)
    {
        log_error("Failed to create GooseSubscriber X");
        GooseReceiver_destroy(receiver);
        return EXIT_FAILURE;
    }
    else
    {
        GooseSubscriber_setDstMac(subscriberX, dstMac);
        GooseSubscriber_setAppId(subscriberX, 1000);
        GooseSubscriber_setListener(subscriberX, gooseListener, NULL);
        GooseReceiver_addSubscriber(receiver, subscriberX);
    }

    GooseReceiver_start(receiver);

    CommParameters gooseCommParameters = {0};
    gooseCommParameters.appId = 1000;
    memcpy(gooseCommParameters.dstAddress, dstMac, 6);
    GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, interface);

    if (publisher == NULL)
    {
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
    // GoosePublisher_setSimulation(publisher, false);
    // GoosePublisher_setNeedsCommission(publisher, false);
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum);

    while (running)
    {
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
