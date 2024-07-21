// logging.c
#include "logging.h"
#include <stdarg.h>

void log_generic(const char *level, const char *format, va_list args) {
    printf("[%s] ", level);
    vprintf(format, args);
    printf("\n");
}

void log_info(const char *format, ...) {
    va_list args;
    va_start(args, format);
    log_generic("INFO", format, args);
    va_end(args);
}

void log_error(const char *format, ...) {
    va_list args;
    va_start(args, format);
    log_generic("ERROR", format, args);
    va_end(args);
}

void log_debug(const char *format, ...) {
    va_list args;
    va_start(args, format);
    log_generic("DEBUG", format, args);
    va_end(args);
}
