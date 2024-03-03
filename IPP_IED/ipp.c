#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <string.h>
#include <unistd.h>
#include <sys/time.h>
#include <time.h>

#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "goose_publisher.h"
#include "hal_thread.h"
#include "mms_value.h"
#include "linked_list.h"

volatile int running = 1;
volatile int ipp_status = 0; // Global variable for the IPP status
static uint32_t stNum = 1;
static uint32_t sqNum = 0;
static char timestamp_str[64]; // Global variable for the timestamp string

// Signal handler for graceful termination
static void sigint_handler(int signalId)
{
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

// GOOSE subscriber listener
static void gooseListener(GooseSubscriber subscriber, void *parameter)
{
    printf("GOOSE event:\n");
    printf("  stNum: %u sqNum: %u\n", GooseSubscriber_getStNum(subscriber),
           GooseSubscriber_getSqNum(subscriber));
    printf("  timeToLive: %u\n", GooseSubscriber_getTimeAllowedToLive(subscriber));

    uint64_t timestamp = GooseSubscriber_getTimestamp(subscriber);

    printf("  timestamp: %u.%u\n", (uint32_t)(timestamp / 1000), (uint32_t)(timestamp % 1000));
    printf("  message is %s\n", GooseSubscriber_isValid(subscriber) ? "valid" : "INVALID");

    MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);

    char buffer[1024];

    MmsValue_printToBuffer(values, buffer, 1024);
    printf("  allData: %s\n", buffer);

    // Update timestamp when a new message is received
    get_utc_timestamp(timestamp_str, sizeof(timestamp_str));

    int new_ipp_status;

    size_t length = strlen(buffer);
    if (buffer[length - 2] == '1')
    {
        new_ipp_status = 0;
    }
    else
    {
        new_ipp_status = 1;
    }

    if (new_ipp_status != ipp_status)
    {
        ipp_status = new_ipp_status;
        stNum++;
        sqNum = 0; // Reset sqNum when stNum changes
    }
}

// GOOSE publisher function
void publish(GoosePublisher publisher)
{
    LinkedList dataSetValues = LinkedList_create();
    char goose_data[100];
    sprintf(goose_data, "IPP Circuit Breaker Status: %s", (ipp_status == 1) ? "CLOSED" : "OPEN");
    LinkedList_add(dataSetValues, MmsValue_newVisibleStringFromByteArray((const uint8_t *)goose_data, strlen(goose_data) + 1));
    // Add the timestamp to the dataset values
    LinkedList_add(dataSetValues, MmsValue_newVisibleString(timestamp_str));
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum++);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1)
    {
        printf("Error sending message!\n");
    }

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

int main(int argc, char **argv)
{
    signal(SIGINT, sigint_handler);

    char *interface = (argc > 1) ? argv[1] : "ens33";
    printf("Using interface %s\n", interface);

    int iterations = 0;
    // Main loop
    while (running)
    {
        // Subscriber setup
        GooseReceiver receiver = GooseReceiver_create();
        GooseReceiver_setInterfaceId(receiver, "ens33");
        GooseSubscriber subscriber = GooseSubscriber_create("simpleIOGenericIO/LLN0$GO$gcbAnalogValues", NULL);
        uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};
        GooseSubscriber_setDstMac(subscriber, dstMac);
        GooseSubscriber_setAppId(subscriber, 1000);
        GooseSubscriber_setListener(subscriber, gooseListener, NULL);
        GooseReceiver_addSubscriber(receiver, subscriber);
        GooseReceiver_start(receiver);

        // Publisher setup
        CommParameters gooseCommParameters = {0};
        gooseCommParameters.appId = 1000;
        memcpy(gooseCommParameters.dstAddress, dstMac, 6);
        GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, "ens33");
        GoosePublisher_setGoCbRef(publisher, "simpleIOGenericIO/LLN0$GO$gcbAnalogValues");
        GoosePublisher_setConfRev(publisher, 1);
        GoosePublisher_setDataSetRef(publisher, "simpleIOGenericIO/LLN0$AnalogValues");
        GoosePublisher_setTimeAllowedToLive(publisher, 500);

        if (GooseReceiver_isRunning(receiver))
        {
            Thread_sleep(100); // Adjust this sleep time as needed
        }

        // Publish the latest status
        publish(publisher);

        Thread_sleep(1000); // Adjust the frequency of publishing as needed
        iterations++;
        if (iterations == 15)
        {
            break;
        }
    }

    return 0;
}
