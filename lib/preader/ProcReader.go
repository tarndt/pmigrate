package preader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"unicode"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib"
	"github.com/tarndt/pmigrate/lib/pfiles"
	"github.com/tarndt/pmigrate/lib/pmaps"
	"github.com/tarndt/pmigrate/lib/ptrace"
)

//Ensure ProcReader implements StateProvider
var _ lib.StateProvider = new(ProcReader)

type ProcReader struct {
	name             string
	process          *ptrace.TracedProcess
	mapFile, memFile *os.File
	backingFileCache map[string]*os.File
	openFiles        []pfiles.FileEntry
}

func NewProcReader(process *os.Process) (*ProcReader, error) {
	//Get process name
	nameBytes, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", process.Pid))
	if err != nil {
		return nil, errs.Append(err, "Could not read process command line.")
	}
	nameBytes = bytes.TrimFunc(nameBytes, func(c rune) bool { return unicode.IsSpace(c) || c == 0 })

	//Attach to target
	tracedProcess, err := ptrace.AttachAndWait(process)
	if err != nil {
		return nil, errs.Append(err, "Could not attach to process: %d, and wait for halt.", process.Pid)
	}

	//Open virtual memory map file
	mapFilePath := fmt.Sprintf("/proc/%d/maps", process.Pid)
	mapFile, err := os.Open(mapFilePath)
	if err != nil {
		return nil, errs.Append(err, "Could not open file: %s, containing target process %d's virtual memory mappings", mapFilePath, process.Pid)
	}

	//Open process memory file
	memFilePath := fmt.Sprintf("/proc/%d/mem", process.Pid)
	memFile, err := os.Open(memFilePath)
	if err != nil {
		return nil, errs.Append(err, "Could not open file: %s, containing target process %d's memory contents", memFilePath, process.Pid)
	}

	//Get list of open files
	openFiles, err := pfiles.GetOpenFiles(process.Pid)
	if err != nil {
		return nil, errs.Append(err, "Could not get a list of open files for target process %d", process.Pid)
	}

	return &ProcReader{
		name:             string(nameBytes),
		process:          tracedProcess,
		mapFile:          mapFile,
		memFile:          memFile,
		backingFileCache: make(map[string]*os.File, 13),
		openFiles:        openFiles,
	}, nil
}

func (this *ProcReader) GetName() string {
	return this.name
}

func (this *ProcReader) GetPID() int {
	return this.process.Pid
}

func (this *ProcReader) GetRegisters() (*syscall.PtraceRegs, error) {
	return this.process.GetRegisters()
}

//Read and parse virtual memory mappings
func (this *ProcReader) GetMemoryMeta() (pmaps.ProcMap, error) {
	var mappings pmaps.ProcMap
	var err error
	if mappings, err = mappings.ParseAppend(this.mapFile); err != nil {
		return nil, errs.Append(err, "Could not parse file: %s, containing target process %d's virtual memory mappings: %s; Details: %s", this.mapFile.Name(), this.process.Pid)
	}
	return mappings, nil
}

func (this *ProcReader) GetMemorySpan(metadata pmaps.Entry) (lib.MemSpan, error) {
	buf := new(bytes.Buffer)
	var err error
	_, err, this.backingFileCache = pmaps.ReadEntry(metadata, this.memFile, buf, true, this.backingFileCache)
	if err != nil {
		return lib.MemSpan{}, errs.Append(err, "Could not read entry: %q", metadata)
	}
	return lib.NewMemSpanReader(metadata, buf), nil
}

func (this *ProcReader) GetFiles() []pfiles.FileEntry {
	return this.openFiles
}

func (this *ProcReader) Close() error {
	this.mapFile.Close()
	this.memFile.Close()
	for _, file := range this.backingFileCache {
		file.Close()
	}
	return this.process.Detach()
}

func (this *ProcReader) GetProcess() *os.Process {
	return this.process.Process
}
