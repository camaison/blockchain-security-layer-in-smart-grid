LIBIEC_HOME=../libiec61850_mod

PROJECT_BINARY_NAME = iedx
PROJECT_SOURCES = iedx.c logging.c  # Added logging.c here

CC=gcc

# Include directories for libiec61850
INCLUDES=-I$(LIBIEC_HOME)/include

# Library paths and names for libiec61850
LDLIBS=-L$(LIBIEC_HOME)/lib -liec61850

# Compiler flags, add any additional flags if needed
CFLAGS=-Wall -std=c99

# Linker flags, add any additional flags if needed
LDFLAGS=

include $(LIBIEC_HOME)/make/target_system.mk
include $(LIBIEC_HOME)/make/stack_includes.mk

all: $(PROJECT_BINARY_NAME)

include $(LIBIEC_HOME)/make/common_targets.mk

$(PROJECT_BINARY_NAME): $(PROJECT_SOURCES) $(LIB_NAME)
	$(CC) $(CFLAGS) $(LDFLAGS) -o $(PROJECT_BINARY_NAME) $(PROJECT_SOURCES) $(INCLUDES) $(LIB_NAME) $(LDLIBS) -lcurl -ljson-c

clean:
	rm -f $(PROJECT_BINARY_NAME)
