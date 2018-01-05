#include "syscall.h"

inline int64 syscall1(int64 syscallNum, int64 arg0) {
	register int64 syscallNum_ __asm__("eax");
	register int64 arg0_ __asm__("edi");
	syscallNum_ = syscallNum;
	arg0_ = arg0;
	asm volatile (
		"syscall"
		: "+r"(syscallNum_)
		: "r"(arg0_)
		: "%rcx"
	);
	return syscallNum_;
}

inline int64 syscall2(int64 syscallNum, int64 arg0, int64 arg1) {
	register int64 syscallNum_ __asm__("eax");
	register int64 arg0_ __asm__("edi");
	register int64 arg1_ __asm__("rsi");
	syscallNum_ = syscallNum;
	arg0_ = arg0;
	arg1_ = arg1;
	asm volatile (
		"syscall"
		: "+r"(syscallNum_)
		: "r"(arg0_), "r"(arg1_)
		: "%rcx", "%r11"
	);
	return syscallNum_;
}

inline int64 syscall3(int64 syscallNum, int64 arg0, int64 arg1, int64 arg2) {
	register int64 syscallNum_ __asm__("eax");
	register int64 arg0_ __asm__("edi");
	register int64 arg1_ __asm__("rsi");
	register int64 arg2_ __asm__("edx");
	syscallNum_ = syscallNum;
	arg0_ = arg0;
	arg1_ = arg1;
	arg2_ = arg2;
	asm volatile (
		"syscall"
		: "+r"(syscallNum_)
		: "r"(arg0_), "r"(arg1_), "r"(arg2_)
		: "%rcx", "%r11", "memory"
	);
	return syscallNum_;
}

inline int64 syscall6(int64 syscallNum, int64 arg0, int64 arg1, int64 arg2,
	int64 arg3, int64 arg4, int64 arg5) {
	register int64 syscallNum_ __asm__("eax");
	register int64 arg0_ __asm__("edi");
	register int64 arg1_ __asm__("rsi");
	register int64 arg2_ __asm__("edx");
	register int64 arg3_ __asm__("r10");
	register int64 arg4_ __asm__("r8");
	register int64 arg5_ __asm__("r9");
	syscallNum_ = syscallNum;
	arg0_ = arg0;
	arg1_ = arg1;
	arg2_ = arg2;
	arg3_ = arg3;
	arg4_ = arg4;
	arg5_ = arg5;
	asm volatile
	(
		"syscall"
		: "+r"(syscallNum_)
		: "r"(arg0_), "r"(arg1_), "r"(arg2_), "r"(arg3), "r"(arg4_), "r"(arg5_) 
		: "%rcx", "%r11", "memory","memory","memory","memory" 
	);
	return syscallNum_;
}

int64 open(char* path, int64 flags, int64 perms) {
	return syscall3(SYS_open, (int64)path, flags, perms);
}

int64 close(int64 fd) {
	return syscall1(SYS_close, fd);
}

inline int64 read(int64 fd, void* buf, int64 len) {
	return syscall3(SYS_read, fd, (int64)buf, len);
}

inline int64 write(int64 fd, void* buf, int64 len) {
	return syscall3(SYS_write, fd, (int64)buf, len);
}

void* mmap(void* addr, int64 len, int64 prot, int64 flags, int64 fd, int64 offset) {
	return (void*)syscall6(SYS_mmap, (int64)addr, len, prot, flags, fd, offset);
}

int64 munmap(void *addr, int64 len) {
	return syscall2(SYS_munmap, (int64)addr, len);
}

int64 mprotect(void* addr, int64 len, int64 prot) {
	return syscall3(SYS_mprotect, (int64)addr, len, prot);
}

void exit(int64 status) {
	syscall1(SYS_exit, status);
}

