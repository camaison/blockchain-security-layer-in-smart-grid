#define _POSIX_C_SOURCE 199309L

#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <signal.h>
#include <time.h>
#include <curl/curl.h>
#include <json-c/json.h>
#include <pthread.h>

#include "mms_value.h"
#include "goose_publisher.h"
#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "hal_thread.h"
#include "logging.h"
#include "hal_time.h"

static volatile int running = 1;
static int rdso_status = 1;
static int update = 0;
static uint32_t stNum = 0;
static uint32_t sqNum = 0;
static char published_timestamp_str[64];
static char subscribed_timestamp_str[64];
static uint32_t subscribed_stNum = 0;
static char subscribed_data[1024] = "FALSE";
static bool statusBool = true;
static char bookkeeping_status[24] = "Valid";

char gocbRef[100] = "RDSO/LLN0$GO$gcbAnalogValues";
char datSet[100] = "RDSO/LLN0$AnalogValues";
char goID[100] = "RDSO";
char goIDListenerIPP[100] = "IPP/LLN0$GO$gcbAnalogValues";
char goIDListenerX[100] = "X/LLN0$GO$gcbAnalogValues";

static pthread_mutex_t lock; // Mutex for thread-safe operations

// Signal handler for graceful termination
static void sigint_handler(int signalId)
{
    running = 0;
}

void log_error_with_retry(const char *message, int retry_count)
{
    fprintf(stderr, "%s Retry count: %d\n", message, retry_count);
}

void formatUtcTime(char *buffer, size_t buffer_size, uint64_t epoch_ms)
{
    time_t rawtime = epoch_ms / 1000;
    struct tm *ptm = gmtime(&rawtime);

    strftime(buffer, buffer_size, "%b %d, %Y %H:%M:%S", ptm);
    // Add milliseconds manually since strftime doesn't support them
    int milliseconds = epoch_ms % 1000;
    char ms_buffer[50];
    sprintf(ms_buffer, ".%03d UTC", milliseconds);
    strcat(buffer, ms_buffer);
}

// Function to send a non-blocking POST request with JSON data
void bookkeeping_api(const char *timestamp, uint32_t stNum, const char *allData, const char *status)
{
    CURL *curl;
    CURLcode res;
    int retry_count = 0;
    int max_retries = 3;
    struct timespec start, end;

    // Record start time
    clock_gettime(CLOCK_REALTIME, &start);

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

            json_object_object_add(jobj, "id", json_object_new_string("RDSO"));
            json_object_object_add(jobj, "status", json_object_new_string(status));
            json_object_object_add(jobj, "message", jmessageContent);

            const char *json_data = json_object_to_json_string(jobj);
            curl_easy_setopt(curl, CURLOPT_URL, "http://192.168.37.139:3001/bookKeeping");
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
            curl_easy_setopt(curl, CURLOPT_TIMEOUT_MS, 5000L); // Set a short timeout

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
            json_object_put(jobj);
        }
    } while (res != CURLE_OK && retry_count < max_retries);

    curl_global_cleanup();

    // Record end time
    clock_gettime(CLOCK_REALTIME, &end);

    // Calculate time difference
    double time_spent = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    printf("Time taken for BookKeeping: %.9f seconds\n", time_spent);
}

void *handle_bookkeeping(void *arg)
{
    bookkeeping_api(published_timestamp_str, stNum, statusBool ? "TRUE" : "FALSE", bookkeeping_status);
    return NULL;
}

// Publish function
void publish(GoosePublisher publisher)
{
    LinkedList dataSetValues = LinkedList_create();
    statusBool = (rdso_status == 1); // True for CLOSED, False for OPEN
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

    // Generate the current timestamp and set it for the publisher
    uint64_t currentTime = Hal_getTimeInMs();
    GoosePublisher_setTimestamp(publisher, currentTime);

    formatUtcTime(published_timestamp_str, sizeof(published_timestamp_str), currentTime);

    if (update)
    {
        update = 0;
        pthread_t update_thread;
        pthread_create(&update_thread, NULL, handle_bookkeeping, NULL);
        pthread_detach(update_thread); // Automatically reclaim thread resources when done
    }

    if (GoosePublisher_publish(publisher, dataSetValues) == -1)
    {
        log_error("Error sending GOOSE message");
    }

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

void gooseListener(GooseSubscriber subscriber, void *parameter)
{
    uint8_t src[6], dst[6];
    GooseSubscriber_getSrcMac(subscriber, src);
    GooseSubscriber_getDstMac(subscriber, dst);
    char srcStr[18], dstStr[18];
    sprintf(srcStr, "%02x:%02x:%02x:%02x:%02x:%02x", src[0], src[1], src[2], src[3], src[4], src[5]);
    sprintf(dstStr, "%02x:%02x:%02x:%02x:%02x:%02x", dst[0], dst[1], dst[2], dst[3], dst[4], dst[5]);

    formatUtcTime(subscribed_timestamp_str, sizeof(subscribed_timestamp_str), GooseSubscriber_getTimestamp(subscriber));

    if (subscribed_stNum != GooseSubscriber_getStNum(subscriber))
    {
        subscribed_stNum = GooseSubscriber_getStNum(subscriber);
    }

    MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
    MmsValue_printToBuffer(values, subscribed_data, sizeof(subscribed_data));
    // Update subscribed_data with a boolean value
    size_t length = strlen(subscribed_data);
    if (length == 6)
    {
        strcpy(subscribed_data, "TRUE");
    }
    else
    {
        strcpy(subscribed_data, "FALSE");
    }

    printf("{\n"
           "\"t\": \"%s\",\n"
           "\"stNum\": %u,\n"
           "\"allData\": %s\n"
           "}\n",
           subscribed_timestamp_str,
           subscribed_stNum,
           subscribed_data);
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
    GooseSubscriber subscriberIPP = GooseSubscriber_create(goIDListenerIPP, NULL);
    GooseSubscriber subscriberX = GooseSubscriber_create(goIDListenerX, NULL);

    uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};

    if (subscriberIPP == NULL)
    {
        log_error("Failed to create GooseSubscriber IPP");
        GooseReceiver_destroy(receiver);
        return EXIT_FAILURE;
    }
    else
    {
        GooseSubscriber_setDstMac(subscriberIPP, dstMac);
        GooseSubscriber_setAppId(subscriberIPP, 1000);
        GooseSubscriber_setListener(subscriberIPP, gooseListener, NULL);
        GooseReceiver_addSubscriber(receiver, subscriberIPP);
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
        GooseSubscriber_destroy(subscriberIPP);
        GooseSubscriber_destroy(subscriberX);
        return EXIT_FAILURE;
    }

    GoosePublisher_setGoCbRef(publisher, gocbRef);
    GoosePublisher_setConfRev(publisher, 1);
    GoosePublisher_setDataSetRef(publisher, datSet);
    GoosePublisher_setTimeAllowedToLive(publisher, 5000);
    GoosePublisher_setGoID(publisher, goID);

    int toggleCounter = 0;
    int count = 0;
    while (count <= 10)
    {
        if (count == 5)
        {
            count++;
            Thread_sleep(26000); // Sleep for 25 seconds
        }
        if (GooseReceiver_isRunning(receiver))
        {
            pthread_mutex_lock(&lock); // Lock the mutex
            if (toggleCounter++ % 5 == 0)
            {
                int old_status = rdso_status;
                rdso_status = !rdso_status;
                if (old_status != rdso_status)
                {
                    stNum++;
                    sqNum = 0;
                    update = 1;
                    count++;
                }
            }
            else
            {
                sqNum++;
            }
            GoosePublisher_setStNum(publisher, stNum);
            GoosePublisher_setSqNum(publisher, sqNum);

            pthread_mutex_unlock(&lock); // Unlock the mutex

            publish(publisher);
            Thread_sleep(1000); // Sleep for 5 seconds
        }
    }

    GoosePublisher_destroy(publisher);
    GooseReceiver_stop(receiver);
    GooseReceiver_destroy(receiver);
    pthread_mutex_destroy(&lock); // Destroy the mutex
    log_info("Application terminated gracefully");

    return 0;
}
