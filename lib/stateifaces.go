package lib

import (
	"bytes"
	"io"
	"io/ioutil"
	"syscall"

	"github.com/tarndt/pmigrate/lib/pfiles"
	"github.com/tarndt/pmigrate/lib/pmaps"
)

type MemSpan struct {
	Metadata pmaps.Entry
	io.ReadCloser
}

func NewMemSpan(metadata pmaps.Entry, rdr io.ReadCloser) MemSpan {
	return MemSpan{Metadata: metadata, ReadCloser: rdr}
}

func NewMemSpanReader(metadata pmaps.Entry, rdr io.Reader) MemSpan {
	return MemSpan{Metadata: metadata, ReadCloser: ioutil.NopCloser(rdr)}
}

func NewMemSpanBytes(metadata pmaps.Entry, data []byte) MemSpan {
	return MemSpan{Metadata: metadata, ReadCloser: ioutil.NopCloser(bytes.NewReader(data))}
}

type StateProvider interface {
	GetName() string
	GetPID() int
	GetRegisters() (*syscall.PtraceRegs, error)
	GetMemoryMeta() (pmaps.ProcMap, error)
	GetMemorySpan(metadata pmaps.Entry) (MemSpan, error)
	GetFiles() []pfiles.FileEntry
	io.Closer
}

type StateConsumer interface {
	Consume(provider StateProvider) error
	DebugInfo() string
	io.Closer
}
