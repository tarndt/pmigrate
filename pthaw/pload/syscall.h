#ifndef SYSCALL_H
#define SYSCALL_H

#include <syscall.h>

#define NULL 0
#define int64 long long
#define uint64 unsigned long long

#define EXIT_SUCCESS 0
#define EXIT_FAILURE 1

#define stdin  0
#define stdout 1
#define stderr 2

//Generic syscall functions 
int64 syscall1(int64 syscallNum, int64 arg0);
int64 syscall2(int64 syscallNum, int64 arg0, int64 arg1);
int64 syscall3(int64 syscallNum, int64 arg0, int64 arg1, int64 arg2);
int64 syscall6(int64 syscallNum, int64 arg0, int64 arg1, int64 arg2, int64 arg3, int64 arg4, int64 arg5);

//Files
int64 open(char* path, int64 flags, int64 perms); //returns file descriptor
int64 close(int64 fd);                            //returns 0 on success, -1 on failure

//IO
int64 read(int64 fd, void* buf, int64 len);  //returns bytes read
int64 write(int64 fd, void* buf, int64 len); //returns bytes written

//mmap
#define PROT_NONE       0x000
#define PROT_READ       0x001
#define PROT_WRITE      0x002
#define PROT_EXEC       0x004

#define MAP_SHARED      0x01
#define MAP_PRIVATE     0x02
#define MAP_FIXED       0x10
#define MAP_ANONYMOUS   0x20

void* mmap(void* addr, int64 len, int64 prot, int64 flags, int64 fd, int64 offset); //returns pointer to allocated block
int64 munmap(void *addr, int64 len);               //returns 0 on success, -1 on failure
int64 mprotect(void* addr, int64 len, int64 prot); //returns 0 on success, -1 on failure

//other
void exit(int64 status);

#endif //SYSCALL_H
