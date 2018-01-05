package pwriter

import (
	"encoding/binary"
	"io"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib"
)

const formatVersion = uint16(1)

//Ensure ProcSnapshotWriter implements StateConsumer
var _ lib.StateConsumer = new(ProcSnapshotWriter)

type ProcSnapshotWriter struct {
	dst io.Writer
}

func NewProcSnapshotWriter(dst io.Writer) *ProcSnapshotWriter {
	return &ProcSnapshotWriter{
		dst: dst,
	}
}

func (this *ProcSnapshotWriter) Consume(provider lib.StateProvider) error {
	const (
		readFailMsg  = "Could not read %q from process state provider"
		writeFailMsg = "Could not write %q to output destination"
	)

	//Before we start writing, get a few items that can fail
	memSpans, err := provider.GetMemoryMeta()
	if err != nil {
		return errs.Append(err, readFailMsg, "memory meta data")
	}
	regs, err := provider.GetRegisters()
	if err != nil {
		return errs.Append(err, readFailMsg, "registers")
	}

	//Format version
	if err = binary.Write(this.dst, binary.LittleEndian, formatVersion); err != nil {
		return errs.Append(err, writeFailMsg, "format version")
	}

	//Process PID
	if err = binary.Write(this.dst, binary.LittleEndian, uint64(provider.GetPID())); err != nil {
		return errs.Append(err, writeFailMsg, "PID")
	}

	//Write name string (with var-bin length prefix)
	buf := make([]byte, binary.MaxVarintLen64)
	name := provider.GetName()
	if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(len(name)))]); err != nil {
		errs.Append(err, writeFailMsg, "process name length")
	}
	if _, err = io.WriteString(this.dst, name); err != nil {
		errs.Append(err, writeFailMsg, "process name")
	}

	//Write registers
	if err = binary.Write(this.dst, binary.LittleEndian, regs); err != nil {
		errs.Append(err, writeFailMsg, "registers")
	}

	//Write open files
	openFiles := provider.GetFiles()
	if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(len(openFiles)))]); err != nil {
		errs.Append(err, writeFailMsg, "open files record count")
	}
	for _, entry := range openFiles {
		//File handle
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(entry.FileHandle))]); err != nil {
			errs.Append(err, writeFailMsg, "open file handle number")
		}
		//File path length, then value
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(len(entry.Path)))]); err != nil {
			errs.Append(err, writeFailMsg, "open file path length")
		}
		if _, err = io.WriteString(this.dst, entry.Path); err != nil {
			return errs.Append(err, writeFailMsg, "open file path value")
		}
		//File type/mode
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(entry.Type))]); err != nil {
			errs.Append(err, writeFailMsg, "open file type/mode")
		}
		//File position
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(entry.Pos))]); err != nil {
			errs.Append(err, writeFailMsg, "open file position")
		}
		//File flags
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(entry.Flags))]); err != nil {
			errs.Append(err, writeFailMsg, "open file position")
		}
	}

	//Write meta-data/data memory span pairs
	for _, entry := range memSpans {
		span, err := provider.GetMemorySpan(entry)
		if err != nil {
			return errs.Append(err, readFailMsg, "memory span")
		}
		//Write span metadata
		entryStr := entry.String()
		if _, err = this.dst.Write(buf[:binary.PutUvarint(buf, uint64(len(entryStr)))]); err != nil {
			return errs.Append(err, writeFailMsg, "span metadata length")
		}
		if _, err = io.WriteString(this.dst, entryStr); err != nil {
			return errs.Append(err, writeFailMsg, "span metadata value")
		}
		if _, err = io.Copy(this.dst, span); err != nil {
			return errs.Append(err, writeFailMsg, "span data")
		}
		span.Close()
	}

	return nil
}

func (this *ProcSnapshotWriter) DebugInfo() string {
	return ""
}

func (this *ProcSnapshotWriter) Close() error {
	return nil
}
