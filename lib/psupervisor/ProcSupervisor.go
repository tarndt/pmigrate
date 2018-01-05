package psupervisor

import (
	"fmt"
	"io"
	"syscall"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib/ptrace"
)

/* TODO:
 * 1. Handle open file descriptors. DONE!
 * 2. Restore TCP sockets?
 * 3. Restore stdin/out?
 */

type ProcSupervisor struct {
	process    *ptrace.TracedProcess
	procStdin  io.Writer
	procStdout io.Reader

	oldPID               uint64
	fileHandleFixupTable map[int]int
}

const verboseDebug = false

func NewProcSupervisor(process *ptrace.TracedProcess, procStdin io.Writer, procStdout io.Reader, oldPID int, fileHandleFixupTable map[int]int) *ProcSupervisor {
	return &ProcSupervisor{
		process:              process,
		procStdin:            procStdin,
		procStdout:           procStdout,
		oldPID:               uint64(oldPID),
		fileHandleFixupTable: fileHandleFixupTable,
	}
}

func (this *ProcSupervisor) Resume() error {
	if err := this.process.Continue(ptrace.NoSignal); err != nil {
		return errs.Append(err, "Could not resume execution of new process")
	}
	return nil
}

//TODO supervise: getpid
func (this *ProcSupervisor) ResumeAndSupervise() error {
	//Supervise forever
	var (
		err             error
		enteringSyscall bool
		syscallCount    int
		syscallID       uint64
		syscallName     string
		registers       syscall.PtraceRegs
	)

	this.process.SetOptionSyscallTraceFlag()

	for {
		enteringSyscall, err = this.process.ContUntilSyscall(ptrace.NoSignal)
		if err != nil {
			return errs.Append(err, "Failure waiting for next syscall, count: %d, kill: %v", syscallCount, this.process.Kill())
		}
		if enteringSyscall {
			syscallCount++
			if err = this.process.GetRegistersInPlace(&registers); err != nil {
				return errs.Append(err, "Failure getting pre-syscall registers, count: %d, kill: %v", syscallCount, this.process.Kill())
			}
			syscallID = registers.Orig_rax
			if verboseDebug {
				syscallName = ptrace.GetSyscallName(syscallID)
				fmt.Printf(" Entering syscall: %s (%d)... ", syscallName, syscallID)
			}
			if err = this.fixSyscallArgs(syscallID, &registers); err != nil {
				return errs.Append(err, "Pre systemcall argument/register fixup failed")
			}
		} else {
			if err = this.process.GetRegistersInPlace(&registers); err != nil {
				return errs.Append(err, "Failure getting pre-syscall registers, count: %d, kill: %v", syscallCount, this.process.Kill())
			}
			if verboseDebug {
				fmt.Printf(" Exited syscall:  %s (%d), result = %d.\n", syscallName, syscallID, registers.Rax)
			}
			if err = this.fixSyscallResults(syscallID, &registers); err != nil {
				return errs.Append(err, "Post systemcall result/register fixup failed")
			}
		}
	}
	return nil
}

func (this *ProcSupervisor) fixSyscallArgs(syscallID uint64, registers *syscall.PtraceRegs) error {
	switch syscallID {
	case syscall.SYS_READ, syscall.SYS_WRITE, syscall.SYS_LSEEK:
		//fmt.Printf("Syscall: %d (%s) made with registers: %+v\n\n", syscallID, ptrace.GetSyscallName(syscallID), registers)
		if newHandle, isPresent := this.fileHandleFixupTable[int(registers.Rdi)]; isPresent {
			//fmt.Printf("Replacing handle: %d with handle %d", registers.Rdi, newHandle)
			registers.Rdi = uint64(newHandle)
			if err := this.process.SetRegisters(registers); err != nil {
				return errs.Append(err, "Could not replace new PID with old PID in RAX syscall result")
			}
			//time.Sleep(time.Second * 30)
			//fmt.Println(" Go!")
		}
	}
	return nil
}

func (this *ProcSupervisor) fixSyscallResults(syscallID uint64, registers *syscall.PtraceRegs) error {
	switch syscallID {
	case syscall.SYS_GETPID:
		registers.Rax = this.oldPID
		if err := this.process.SetRegisters(registers); err != nil {
			return errs.Append(err, "Could not replace new PID with old PID in RAX syscall result")
		}
	}
	return nil
}
