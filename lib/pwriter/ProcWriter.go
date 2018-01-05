package pwriter

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib"
	"github.com/tarndt/pmigrate/lib/pfiles"
	"github.com/tarndt/pmigrate/lib/psupervisor"
	"github.com/tarndt/pmigrate/lib/ptrace"
)

const (
	opStart   = 65
	opMemLoad = 66
	opExec    = 67
	opAbort   = 68

	respStarted   = 97
	respMemloaded = 98
	respExecing   = 99
	respAborting  = 100
	respFail      = 101
)

//Ensure ProcWriter implements StateConsumer
var _ lib.StateConsumer = new(ProcWriter)

type StdioSinks struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func DefaultStdioSinks() StdioSinks {
	return StdioSinks{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

type ProcWriter struct {
	loaderPath string
	stdioSinks StdioSinks

	ldr                  *exec.Cmd
	ldrIn                *bufio.Writer
	ldrOut               *bufio.Reader
	fileHandleFixupTable map[int]int //Old file # -> new file #, used by supervisor to fixup system calls
}

func NewProcWriter(loaderPath string) *ProcWriter {
	return NewProcWriterCustStdio(loaderPath, DefaultStdioSinks())
}

func NewProcWriterCustStdio(loaderPath string, stdioSinks StdioSinks) *ProcWriter {
	return &ProcWriter{loaderPath: loaderPath, stdioSinks: stdioSinks}
}

func (this *ProcWriter) Consume(provider lib.StateProvider) error {
	regs, err := provider.GetRegisters()
	if err != nil {
		return errs.Append(err, "Could not get registers")
	}
	spans, err := provider.GetMemoryMeta()
	if err != nil {
		return errs.Append(err, "Could not get memory metadata")
	}
	//Startup external loader
	if err = this.start(provider.GetFiles()); err != nil {
		return err
	}
	//Send memory mappings to loader
	for _, spanMeta := range spans {
		span, err := provider.GetMemorySpan(spanMeta)
		if err != nil {
			return errs.Append(err, "Could not get memory span")
		}
		if err = this.sendSpan(span); err != nil {
			return err
		}
		span.Close()
	}
	//Start execution
	if err = this.run(regs, provider.GetPID()); err != nil {
		return err
	}
	this.abort()
	return nil
}

func (this *ProcWriter) DebugInfo() string {
	return ""
}

func (this *ProcWriter) Close() error {
	return nil
}

func (this *ProcWriter) start(openFiles []pfiles.FileEntry) error {
	//Start loader
	var err error
	if this.ldr, this.ldrIn, this.ldrOut, this.fileHandleFixupTable, err = startLoader(this.loaderPath, openFiles, this.stdioSinks); err != nil {
		return err
	}
	//Send startup ack
	if err = this.ldrIn.WriteByte(opStart); err != nil {
		return errs.Append(err, "Could not send command: ", opStart)
	}
	if err = this.ldrIn.Flush(); err != nil {
		return err
	}
	if err = checkResp(this.ldrOut, respStarted); err != nil {
		return err
	}
	return nil
}

func (this *ProcWriter) sendSpan(span lib.MemSpan) error {
	if span.Metadata.FileInfo.Path() == "[vsyscall]" {
		return nil
	}
	//Send command
	err := this.ldrIn.WriteByte(opMemLoad)
	if err != nil {
		return errs.Append(err, "Could not send command: ", opMemLoad)
	}
	this.ldrIn.Flush()
	//Send mmap args
	const argErr = "Could not send memory span metadata"
	if err = binary.Write(this.ldrIn, binary.LittleEndian, int64(span.Metadata.MemStart)); err != nil {
		errs.Append(err, argErr)
	}
	if err = binary.Write(this.ldrIn, binary.LittleEndian, int64(span.Metadata.Len())); err != nil {
		errs.Append(err, argErr)
	}
	if err = binary.Write(this.ldrIn, binary.LittleEndian, span.Metadata.Perms.Cvalue()); err != nil {
		errs.Append(err, argErr)
	}
	if _, err = this.ldrIn.ReadFrom(span); err != nil {
		return errs.Append(err, "Could not send memory span data for entry: %s", span.Metadata)
	}
	if err = this.ldrIn.Flush(); err != nil {
		return err
	}
	//Check response
	if err := checkResp(this.ldrOut, respMemloaded); err != nil {
		return err
	}
	return nil
}

func (this *ProcWriter) abort() error {
	err := this.ldrIn.WriteByte(opStart)
	if err != nil {
		return errs.Append(err, "Could not send command: ", opAbort)
	}
	this.ldrIn.Flush()
	if err = checkResp(this.ldrOut, respAborting); err != nil {
		return err
	}
	return nil
}

func (this *ProcWriter) run(regs *syscall.PtraceRegs, oldPID int) error {
	//Send command
	err := this.ldrIn.WriteByte(opExec)
	if err != nil {
		return errs.Append(err, "Could not send command: ", opMemLoad)
	}
	this.ldrIn.Flush()
	if err = checkResp(this.ldrOut, respExecing); err != nil {
		return err
	}
	//Ptrace
	os.Stderr.WriteString("Attaching... ")
	ldr, err := ptrace.AttachAndWait(this.ldr.Process)
	if err != nil {
		return errs.Append(err, "Could not attach to loader process: %d, and wait for halt.", this.ldr.Process.Pid)
	}
	os.Stderr.WriteString("Attached.\n")
	//Load registers
	os.Stderr.WriteString("Loading Registers... ")
	if err = ldr.SetRegisters(regs); err != nil {
		return errs.Append(err, "Could not load registers into new process")
	}
	os.Stderr.WriteString("Loaded.\n")
	//Resume process, process should be restored!
	os.Stderr.WriteString("Resuming process... \n")

	supervisor := psupervisor.NewProcSupervisor(ldr, this.ldrIn, this.ldrOut, oldPID, this.fileHandleFixupTable)
	return errs.Append(supervisor.ResumeAndSupervise(), "Process supervision failed")
}

func checkResp(src io.ByteReader, expected byte) error {
	if resp, err := src.ReadByte(); err != nil {
		return errs.Append(err, "Could not read response from stream")
	} else if resp != expected {
		return errs.Append(err, "Unexpected response from loader; expected: %d, received: %d", expected, resp)
	}
	return nil
}

func startLoader(loaderPath string, openFiles []pfiles.FileEntry, stdioSinks StdioSinks) (*exec.Cmd, *bufio.Writer, *bufio.Reader, map[int]int, error) {
	toLoaderRdr, toLoaderWtr, err := os.Pipe()
	if err != nil {
		return nil, nil, nil, nil, errs.Append(err, "Could not create pipe 1 (to-loader) to communicate with loader")
	}
	toParentRdr, toParentWtr, err := os.Pipe()
	if err != nil {
		return nil, nil, nil, nil, errs.Append(err, "Could not create pipe 2 (to-supervisor) to communicate with loader")
	}

	cmd := exec.Command(loaderPath)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdioSinks.Stdin, stdioSinks.Stdout, stdioSinks.Stderr
	cmd.ExtraFiles = []*os.File{toLoaderRdr, toParentWtr}

	//Setup any open files that need to be restored
	fileHandleFixupTable := make(map[int]int, len(openFiles)) // Old file # -> new file #
	curFdPos := 5                                             //3 standard files + our two pipes above = 4
	for _, entry := range openFiles {
		if !entry.Type.IsRegular() { //We only try to restore 'regular files' for now
			continue
		}
		file, err := os.OpenFile(entry.Path, entry.Flags, 0)
		if err != nil {
			return nil, nil, nil, nil, errs.Append(err, "Failure to open file while attempting to restore file: %s", entry)
		}
		if _, err = file.Seek(int64(entry.Pos), os.SEEK_SET); err != nil {
			return nil, nil, nil, nil, errs.Append(err, "Failure to returning to last seek postion in open file while attempting to restore file: %s", entry)
		}

		cmd.ExtraFiles = append(cmd.ExtraFiles, file)
		//This table is used later to fixup system calls
		fileHandleFixupTable[entry.FileHandle] = curFdPos
		curFdPos++
	}

	//Start loader execution
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, nil, errs.Append(err, "Could not execute loader at path: %s", loaderPath)
	}
	return cmd, bufio.NewWriter(toLoaderWtr), bufio.NewReader(toParentRdr), fileHandleFixupTable, nil
}
