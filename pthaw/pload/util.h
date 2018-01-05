#ifndef UTIL_H
#define UTIL_H

#include "syscall.h"

typedef enum { false, true } bool;

int64 puts(char* msg);
int64 fputs(char* msg, int64 fd);
int64 strlen(char* msg);

#endif //UTIL_H
