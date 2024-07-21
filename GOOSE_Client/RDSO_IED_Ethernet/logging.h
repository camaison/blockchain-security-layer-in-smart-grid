// logging.h
#ifndef LOGGING_H
#define LOGGING_H

#include <stdio.h>

void log_info(const char *format, ...);
void log_error(const char *format, ...);
void log_debug(const char *format, ...);

#endif // LOGGING_H
