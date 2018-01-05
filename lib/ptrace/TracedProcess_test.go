package ptrace

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	sleepCmd    string
	listDirCmd  string
	hostnameCmd string
)

func init() {
	sleepCmd = mustCmdFromPath("sleep")
	hostnameCmd = mustCmdFromPath("hostname")
}

func mustCmdFromPath(cmd string) string {
	output, err := exec.Command("which", cmd).Output()
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimSpace(output))
}

func TestAttaching(t *testing.T) {
	tracedProcess := startProcessAttach(t, sleepCmd, "1")
	if err := tracedProcess.Continue(NoSignal); err != nil {
		t.Fatalf("TracedProcess.Continue() returned: %s", err)
	}
}

func TestGetRegisters(t *testing.T) {
	tracedProcess := startProcessAttach(t, sleepCmd, "5")
	if _, err := tracedProcess.GetRegisters(); err != nil {
		t.Fatalf("TracedProcess.GetRegisters() returned: %s", err)
	}
}

func TestSyscall(t *testing.T) {
	//TODO
	// 1. Call hostname, get hostname
	//const cmd = "hostname"
	//hostname, err := exec.Command(cmd).Output()
	//if err != nil {
	//	t.Fatalf("Could not call: %s; Details: %s", cmd, err)
	//}
	// 2. Call hostname again, attach, intercept "uname" replace with "INTERCEPTED"
	//tracedProcess := startProcessAttach(t, hostnameCmd)
	//tracedProcess.Syscall(syscall.Uname)
}

func startProcessAttach(t *testing.T, cmdPath string, args ...string) *TracedProcess {
	pargs := make([]string, 1, len(args)+1)
	pargs[0] = filepath.Base(cmdPath)
	pargs = append(pargs, args...)
	process, err := os.StartProcess(sleepCmd, pargs, new(os.ProcAttr))
	if err != nil {
		t.Fatalf("Attaching to: %s, failed; Details: %s", sleepCmd, err)
	}
	tracedProcess, err := Attach(process)
	if err != nil {
		t.Fatalf("TracedProcess.Attach(%v) returned: %s", process, err)
	}
	if err = tracedProcess.WaitStopped(); err != nil {
		t.Fatalf("TracedProcess.WaitStopped() returned: %s", err)
	}
	return tracedProcess
}
