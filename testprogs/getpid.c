#include <stdio.h>
#include <stdbool.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <syscall.h>

const unsigned int annouceEveryXSecs = 10;

int main() {
	pid_t myPID = syscall(SYS_getpid); // getpid(); Don't use this glibc caches...
	pid_t parentPID = getppid();
	printf("This process's PID is %d.\n", myPID);
	printf("This process's parent's PID is %d.\n\n", parentPID);

	unsigned int secsSinceChange = 0;
	while(true) {
		pid_t newPID = syscall(SYS_getpid); //See comment above
		pid_t newParentPID = getppid();
		if(myPID != newPID) {
			printf("This process's PID changed from %d to %d!\n", myPID, newPID);
			myPID = newPID;
			secsSinceChange	= 0;	
		}
		if(parentPID != newParentPID) {
			printf("This process's parent's PID changed from %d to %d!\n", parentPID, newParentPID);
			parentPID = newParentPID;
			secsSinceChange	= 0;		
		}

		if(secsSinceChange > 0 && secsSinceChange % annouceEveryXSecs == 0) {
			printf("No changes in PID or PPID in last %d seconds.\n", annouceEveryXSecs);			
		}
		sleep(1);
		secsSinceChange++;		
	}
	return EXIT_SUCCESS;
}

