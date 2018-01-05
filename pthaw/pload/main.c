#include "util.h"

void execByteCode();
void ack(char respCode);

void main() {
	execByteCode();
	exit(EXIT_SUCCESS);
}

//Constants for open file handles passed by parent supervisor. We use these for
// communicating wih out parent.
#define ldrIn  3
#define ldrOut 4


#define opStart   65
#define opMemLoad 66
#define opExec    67
#define opAbort   68

#define respStarted    97
#define respMemloaded  98
#define respExecing    99
#define respAborting  100
#define respFail      101

void execByteCode() {
	char opCode;
	int64 mmapArgs[3]; //6 - 3 = 3, we ignore flags, fd and offset
	
	//Buffer used for memory xfers
	const int64 bufLen = 512;
	char buf[bufLen];
	int64 n;
	
	bool loop = true;
	while(loop) {
		//Read operation code
		if(read(ldrIn, &opCode, sizeof(opCode)) != 1) {
			fputs("Error: Could not read command code!\n", stderr);
			exit(EXIT_FAILURE);
		}
		switch(opCode) {
			case opMemLoad:
				//Read mmap args
				if(read(ldrIn, &mmapArgs, sizeof(mmapArgs)) != sizeof(mmapArgs)) {
					fputs("Error: Could not read arguments for mmap operation!\n", stderr);
					exit(EXIT_FAILURE);
				}
				//mmap, but always allow write so we can populate it
				void* addr = (void*)mmapArgs[0];
				int64 len = mmapArgs[1];
				int64 prot = mmapArgs[2]; 
				if(mmap(addr, len, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS|MAP_FIXED, -1, 0) != addr) {
					fputs("Error: Failed to create mapping at correct address!\n", stderr); //Did you send me a [vsyscall] line?
					exit(EXIT_FAILURE);
				}
				//copy memory contents in
				char* baseAddr = addr;
				int64 cur = 0, bufPos = 0, bufEnd = 0;
				while(cur < len) {
					while(bufPos < bufEnd) {
						baseAddr[cur] = buf[bufPos];
						cur++;
						bufPos++;
					}
					//Get next batch
					if(len - cur > bufLen) {
						bufEnd = bufLen;
					} else {
						bufEnd = len - cur;
					}
					if(read(ldrIn, &buf, bufEnd) != bufEnd) {
						fputs("Error: Could not populate memory contents!\n", stderr);
						exit(EXIT_FAILURE);
					}
					bufPos = 0;					
				}
				//If the memory mapping protection we used for loading was not the
				//one used by the application being restored, set it correctly.
				if(prot != PROT_READ|PROT_WRITE && mprotect(addr, len, prot) != 0) {
					fputs("Error: Could not set orginal protections on memory mapping!\n", stderr);
					exit(EXIT_FAILURE);
				}
				ack(respMemloaded);
				continue;			
			case opStart:
				//Used to sanity check we are getting a valid data-stream
				ack(respStarted);
				continue;
			case opExec:
				//Send a ready message to the parent, after which we busy wait while
				// waiting for the parent to ptrace, load registers and resume
				// execution (with loaded code). Execution ends here.
				ack(respExecing);
				while(true) {
					fputs(".", stderr);
				};				
			case opAbort:
				ack(respAborting);
				loop = false;
				break;
			default:
				ack(respFail);
				fputs("Error: Unknown opcode, execution aborted!\n", stderr);
				exit(EXIT_FAILURE);				
		}
	}
}

void ack(char respCode) {
	if(write(ldrOut, &respCode, sizeof(respCode)) != sizeof(respCode)) {
		fputs("Error: Could not send response status code!\n", stderr);
		exit(EXIT_FAILURE);	
	}
}

