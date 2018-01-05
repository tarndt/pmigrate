package ptrace

import (
	"errors"
	"os"
	"syscall"

	"lib/errs"
)

const NoSignal syscall.Signal = 0

type TracedProcess struct {
	*os.Process
	inSyscall bool
}

var errStatusNotStopped = errors.New("Process was was not stopped")

func Attach(process *os.Process) (*TracedProcess, error) {
	err := syscall.PtraceAttach(process.Pid)
	switch {
	case err == syscall.EPERM:
		if _, ptraceErr := syscall.PtraceGetEventMsg(process.Pid); ptraceErr != nil {
			return nil, ptraceErr
		}
		fallthrough
	case err != nil:
		return nil, errs.Append(err, "Could not attach")
	}
	return &TracedProcess{Process: process, inSyscall: true}, nil
}

func AttachAndWait(process *os.Process) (*TracedProcess, error) {
	tracedProcess, err := Attach(process)
	if err != nil {
		return nil, err
	}
	return tracedProcess, tracedProcess.WaitStopped()
}

func (this *TracedProcess) WaitStopped() error {
	status, err := this.WaitStatus()
	if err == nil && !status.Stopped() {
		err = errStatusNotStopped
	}
	return err
}

func (this *TracedProcess) WaitStatus() (syscall.WaitStatus, error) {
	status := syscall.WaitStatus(0)
	_, err := syscall.Wait4(this.Pid, &status, 0, nil)
	return status, err
}

func (this *TracedProcess) Detach() error {
	return syscall.PtraceDetach(this.Pid)
}

func (this *TracedProcess) SetOptions(options int) error {
	return syscall.PtraceSetOptions(this.Pid, options)
}

func (this *TracedProcess) SetOptionSyscallTraceFlag() error {
	const PTRACE_O_TRACESYSGOOD = 1
	return this.SetOptions(PTRACE_O_TRACESYSGOOD)
}

func (this *TracedProcess) GetEventMsg() (uint, error) {
	return syscall.PtraceGetEventMsg(this.Pid)
}

func (this *TracedProcess) Continue(signal syscall.Signal) error {
	return syscall.PtraceCont(this.Pid, int(signal))
}

func (this *TracedProcess) SingleStep() error {
	return syscall.PtraceSingleStep(this.Pid)
}

//Syscall continues the child process and halts on the next enterence or exit to a system call
func (this *TracedProcess) ContUntilSyscall(signal syscall.Signal) (bool, error) {
	const SYSCALL_TRAP = syscall.SIGTRAP | 0x80

	for {
		err := syscall.PtraceSyscall(this.Pid, int(signal))
		if err != nil {
			if err == syscall.ESRCH {
				if isAlive(this.Pid) {
					continue
				}
			}
			return false, errs.Append(err, "PTRACE_SYSCALL operation failed unexpectedly")
		}

		var status syscall.WaitStatus
		if status, err = this.WaitStatus(); err != nil {
			return false, errs.Append(err, "Waiting for process to enter/exit syscall failed")
		}

		switch {
		case status.Exited():
			return false, errs.Append(err, "Process exited!; return code was: %d", status.ExitStatus())
		case !status.Stopped():
			return false, errs.Append(err, "Process was not stopped; status is: %X", status)
		case status.StopSignal() != SYSCALL_TRAP:
			//continue
			return false, errs.Append(err, "Process stopped for reason other than a syscall SIGTRAP; reason was: %X", status.StopSignal())
		}
		break //We must be in a stopped state with a status of SIGTRAP
	}
	curSyscallState := this.inSyscall
	this.inSyscall = !this.inSyscall
	return curSyscallState, nil
}

func isAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	return err == nil && proc.Signal(syscall.Signal(0)) != syscall.ESRCH
}

func (this *TracedProcess) GetRegisters() (*syscall.PtraceRegs, error) {
	registers := new(syscall.PtraceRegs)
	return registers, this.GetRegistersInPlace(registers)
}

func (this *TracedProcess) GetRegistersInPlace(registers *syscall.PtraceRegs) error {
	return syscall.PtraceGetRegs(this.Pid, registers)
}

func (this *TracedProcess) SetRegisters(registers *syscall.PtraceRegs) error {
	return syscall.PtraceSetRegs(this.Pid, registers)
}

func (this *TracedProcess) PeekData(address uintptr, out []byte) (int, error) {
	return syscall.PtracePeekData(this.Pid, address, out)
}

func (this *TracedProcess) PokeData(address uintptr, data []byte) (int, error) {
	return syscall.PtracePokeData(this.Pid, address, data)
}

func (this *TracedProcess) PeekText(address uintptr, out []byte) (int, error) {
	return syscall.PtracePeekText(this.Pid, address, out)
}

func (this *TracedProcess) PokeText(address uintptr, data []byte) (int, error) {
	return syscall.PtracePokeText(this.Pid, address, data)
}
