#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <signal.h>

#include "mms_value.h"
#include "goose_publisher.h"
#include "goose_receiver.h"
#include "goose_subscriber.h"
#include "hal_thread.h"

static volatile int running = 1;
static int rdso_status = 1;
static uint32_t stNum = 1;
static uint32_t sqNum = 0;

// Signal handler for graceful termination
static void sigint_handler(int signalId)
{
    running = 0;
}

void publish(GoosePublisher publisher)
{
    LinkedList dataSetValues = LinkedList_create();
    char goose_data[100];
    sprintf(goose_data, "RDSO Circuit Breaker Status: %s", (rdso_status == 1) ? "CLOSED" : "OPEN");
    LinkedList_add(dataSetValues, MmsValue_newVisibleStringFromByteArray((const uint8_t *)goose_data, strlen(goose_data) + 1));
    LinkedList_add(dataSetValues, MmsValue_newIntegerFromInt32(rdso_status));
    GoosePublisher_setStNum(publisher, stNum);
    GoosePublisher_setSqNum(publisher, sqNum++);

    if (GoosePublisher_publish(publisher, dataSetValues) == -1)
    {
        printf("Error sending message!\n");
    }

    LinkedList_destroyDeep(dataSetValues, (LinkedListValueDeleteFunction)MmsValue_delete);
}

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
}

int main(int argc, char **argv)
{

    signal(SIGINT, sigint_handler);

    char *interface = (argc > 1) ? argv[1] : "ens33";
    printf("Using interface %s\n", interface);
    // Main loop
    int toggleCounter = 0;
    int count = 0;

    while (running && count < 15)
    {
        // Publisher setup
        CommParameters gooseCommParameters = {0};
        gooseCommParameters.appId = 1000;
        gooseCommParameters.dstAddress[0] = 0x01;
        gooseCommParameters.dstAddress[1] = 0x0c;
        gooseCommParameters.dstAddress[2] = 0xcd;
        gooseCommParameters.dstAddress[3] = 0x01;
        gooseCommParameters.dstAddress[4] = 0x00;
        gooseCommParameters.dstAddress[5] = 0x01;
        GoosePublisher publisher = GoosePublisher_create(&gooseCommParameters, interface);
        GoosePublisher_setGoCbRef(publisher, "simpleIOGenericIO/LLN0$GO$gcbAnalogValues");
        GoosePublisher_setConfRev(publisher, 1);
        GoosePublisher_setDataSetRef(publisher, "simpleIOGenericIO/LLN0$AnalogValues");
        GoosePublisher_setTimeAllowedToLive(publisher, 500);

        // Subscriber setup
        GooseReceiver receiver = GooseReceiver_create();
        GooseReceiver_setInterfaceId(receiver, interface);
        GooseSubscriber subscriber = GooseSubscriber_create("simpleIOGenericIO/LLN0$GO$gcbAnalogValues", NULL);
        uint8_t dstMac[6] = {0x01, 0x0c, 0xcd, 0x01, 0x00, 0x01};
        GooseSubscriber_setDstMac(subscriber, dstMac);
        GooseSubscriber_setAppId(subscriber, 1000);
        GooseSubscriber_setListener(subscriber, gooseListener, NULL);
        GooseReceiver_addSubscriber(receiver, subscriber);
        GooseReceiver_start(receiver);

        if (GooseReceiver_isRunning(receiver))
        {
            Thread_sleep(100); // Adjust this sleep time as needed
        }

        // Toggle the rdso_status every 10 iterations
        if (toggleCounter++ % 10 == 0)
        {
            int old_status = rdso_status;
            rdso_status = (rdso_status == 1) ? 0 : 1;
            if (old_status != rdso_status)
            {
                stNum++;
                sqNum = 0; // Reset sqNum when stNum changes
            }
            toggleCounter = 0;
        }

        // Publish the latest status
        publish(publisher);
        count++;
        Thread_sleep(1000); // Adjust the frequency of publishing as needed

        // Clean up
        GoosePublisher_destroy(publisher);
        GooseReceiver_stop(receiver);
        GooseReceiver_destroy(receiver);
    }
    return 0;
}
