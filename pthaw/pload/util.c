#include "util.h"

int64 puts(char* msg) {
	return write(stdout, (void*)msg, strlen(msg));
}

int64 fputs(char* msg, int64 fd) {
	return write(fd, (void*)msg, strlen(msg));
}

int64 strlen(char* msg) {
	if(msg == 0) {
		return 0;
	}
	char* msgStart = msg;
	for(; *msg != 0; msg++);
	return msg-msgStart;	
}
