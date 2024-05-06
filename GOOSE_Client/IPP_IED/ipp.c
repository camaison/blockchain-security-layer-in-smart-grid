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
#include "logging.h"  // Custom logging library

volatile int running = 1;
volatile int ipp_status = 0; // Global variable for the IPP status
static uint32_t stNum = 0;
static uint32_t sqNum = 0;

// Signal handler for graceful termination
static void sigint_handler(int signalId) {
    running = 0;
}

// GOOSE subscriber listener
// static void gooseListener(GooseSubscriber subscriber, void *parameter) {
//     log_info("GOOSE event: stNum: %u, sqNum: %u, timeToLive: %u", 
//              GooseSubscriber_getStNum(subscriber), GooseSubscriber_getSqNum(subscriber),
//              GooseSubscriber_getTimeAllowedToLive(subscriber));

//     uint64_t timestamp = GooseSubscriber_getTimestamp(subscriber);
//     log_info("timestamp: %u.%u", (uint32_t)(timestamp / 1000), (uint32_t)(timestamp % 1000));
//     log_info("message is %s", GooseSubscriber_isValid(subscriber) ? "valid" : "INVALID");

//     MmsValue *values = GooseSubscriber_getDataSetValues(subscriber);
//     char buffer[1024];
//     MmsValue_printToBuffer(values, buffer, 1024);
//     log_info("allData: %s", buffer);

//     // Update timestamp when a new message is received
//     get_utc_timestamp(timestamp_str, sizeof(timestamp_str));

//     int new_ipp_status = buffer[strlen(buffer) - 2] == '1' ? 0 : 1;
//     if (new_ipp_status != ipp_status) {
//         ipp_status = new_ipp_status;
//         stNum++;
//         sqNum = 0; // Reset sqNum when stNum changes
//     }
// }
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

    printf("{\n"
           "\"goID\": \"%s\",\n"
           "\"Src\": \"%s\",\n"
           "\"Dst\": \"%s\",\n"
           "\"t\": \"%s\",\n"
           "\"size\": %d,\n"
           "\"gocbRef\": \"%s\",\n"
           "\"stNum\": %u,\n"
           "\"allData\": %s\n"
           "}\n",
           GooseSubscriber_getGoId(subscriber),
           srcStr, dstStr,
           formattedTime,
           189, // This should be dynamic, based on actual data size
           GooseSubscriber_getGoCbRef(subscriber),
           GooseSubscriber_getStNum(subscriber),
           GooseSubscriber_getDataSetValues(subscriber) ? "TRUE" : "FALSE");
}


// GOOSE publisher function
void publish(GoosePublisher publisher) {
    LinkedList dataSetValues = LinkedList_create();
    bool statusBool = (ipp_status == 1);
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));
    // GoosePublisher_setStNum(publisher, stNum);
    // GoosePublisher_setSqNum(publisher, sqNum++);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
        log_error("Error sending GOOSE message");
    } else {
        log_debug("GOOSE message sent successfully: Status %s", statusBool ? "True" : "False");
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
    GoosePublisher_setTimeAllowedToLive(publisher, 500);

    int toggleCounter = 0;
    while (running) {
        if (GooseReceiver_isRunning(receiver)) {
            if (toggleCounter++ % 10 == 0) {
                int old_status = ipp_status;
                ipp_status = !ipp_status;
                if (old_status != ipp_status) {
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
