package preader

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"syscall"

	"lib"
	"lib/errs"
	"lib/pfiles"
	"lib/pmaps"
)

const formatVersion = uint16(1)

//Ensure ProcSnapReader implements StateProvider
var _ lib.StateProvider = new(ProcSnapReader)

type flexReader interface {
	io.Reader
	io.ByteReader
}

type ProcSnapReader struct {
	name      string
	pid       uint64
	regs      syscall.PtraceRegs
	memMeta   pmaps.ProcMap
	memData   map[uint64]lib.MemSpan
	openFiles []pfiles.FileEntry
}

func NewProcSnapReader(inStrm flexReader) (*ProcSnapReader, error) {
	const readFailMsg = "Could not read %q from process snapshot stream"

	this := &ProcSnapReader{
		memData: make(map[uint64]lib.MemSpan, 31),
	}

	//Format version
	var fmtVer uint16
	err := binary.Read(inStrm, binary.LittleEndian, &fmtVer)
	if err != nil {
		return nil, errs.Append(err, readFailMsg, "format version")
	} else if fmtVer < formatVersion {
		return nil, errs.New("Unsupported format version, snapshot was version %d, and this tool only understands up to: %d", fmtVer, formatVersion)
	}

	//PID
	if err := binary.Read(inStrm, binary.LittleEndian, &this.pid); err != nil {
		return nil, errs.Append(err, readFailMsg, "PID")
	}

	//Name
	if this.name, err = getStr(inStrm); err != nil {
		return nil, errs.Append(err, readFailMsg, "process name")
	}

	//Registers
	if err = binary.Read(inStrm, binary.LittleEndian, &this.regs); err != nil {
		return nil, errs.Append(err, readFailMsg, "registers")
	}

	//Open files
	var temp uint64
	if temp, err = binary.ReadUvarint(inStrm); err != nil {
		return nil, errs.Append(err, readFailMsg, "open files record count")
	}
	openFileCount := int(temp)
	this.openFiles = make([]pfiles.FileEntry, openFileCount)
	for i := 0; i < openFileCount; i++ {
		entry := &this.openFiles[i]
		//File handle
		if temp, err = binary.ReadUvarint(inStrm); err != nil {
			return nil, errs.Append(err, readFailMsg, "open file handle number")
		}
		entry.FileHandle = int(temp)
		//File path
		if entry.Path, err = getStr(inStrm); err != nil {
			return nil, errs.Append(err, readFailMsg, "open file path")
		}
		//File type/mode
		if temp, err = binary.ReadUvarint(inStrm); err != nil {
			return nil, errs.Append(err, readFailMsg, "open file type/mode")
		}
		entry.Type = os.FileMode(temp)
		//File position
		if temp, err = binary.ReadUvarint(inStrm); err != nil {
			return nil, errs.Append(err, readFailMsg, "open file handle number")
		}
		entry.Pos = int(temp)
		//File flags
		if temp, err = binary.ReadUvarint(inStrm); err != nil {
			return nil, errs.Append(err, readFailMsg, "open file handle number")
		}
		entry.Flags = int(temp)
	}

	//Read meta-data/data memory span pairs
	var buf bytes.Buffer
	for {
		buf.Reset()
		if err = getStrBuf(inStrm, &buf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errs.Append(err, readFailMsg, "span metadata")
		}
		var metadata pmaps.Entry
		metadata, err = pmaps.ParseEntry(&buf)
		if err != nil {
			return nil, errs.Append(err, "Could not parse span metadata")
		}
		data := make([]byte, metadata.Len())
		if _, err = io.ReadFull(inStrm, data); err != nil {
			return nil, errs.Append(err, readFailMsg, "span data")
		}
		this.addMemSpan(metadata, data)
	}
	return this, nil
}

func getStrBuf(rdr flexReader, buf *bytes.Buffer) error {
	strLen, err := binary.ReadUvarint(rdr)
	if err != nil {
		return err
	}
	buf.Grow(int(strLen))
	_, err = io.CopyN(buf, rdr, int64(strLen))
	return err
}

func getStr(rdr flexReader) (string, error) {
	var buf bytes.Buffer
	err := getStrBuf(rdr, &buf)
	return buf.String(), err
}

func (this *ProcSnapReader) GetName() string {
	return this.name
}

func (this *ProcSnapReader) GetPID() int {
	return int(this.pid)
}

func (this *ProcSnapReader) GetRegisters() (*syscall.PtraceRegs, error) {
	return &this.regs, nil
}

func (this *ProcSnapReader) GetMemoryMeta() (pmaps.ProcMap, error) {
	return this.memMeta, nil
}

func (this *ProcSnapReader) GetMemorySpan(metadata pmaps.Entry) (lib.MemSpan, error) {
	if span, isPresent := this.memData[metadata.MemStart]; isPresent {
		return span, nil
	}
	return lib.MemSpan{}, errs.New("Memory span at start address: %d, does not exist", metadata.MemStart)
}

func (this *ProcSnapReader) GetFiles() []pfiles.FileEntry {
	return this.openFiles
}

func (this *ProcSnapReader) Close() error {
	return nil
}

func (this *ProcSnapReader) addMemSpan(metadata pmaps.Entry, data []byte) {
	this.memMeta = append(this.memMeta, metadata)
	memStart := metadata.MemStart
	spanRdr := memSpan{
		memStart: memStart,
		memData:  this.memData,
		Reader:   bytes.NewReader(data),
	}
	span := lib.NewMemSpan(metadata, spanRdr)
	this.memData[memStart] = span
}

type memSpan struct {
	memStart uint64
	memData  map[uint64]lib.MemSpan
	io.Reader
}

func (this memSpan) Close() error {
	this.Reader = nil
	delete(this.memData, this.memStart)
	return nil
}
