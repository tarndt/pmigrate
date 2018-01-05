package pmaps

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"lib/errs"
)

var ErrInvalidMemorySpan = errors.New("Invalid memory span: End must be >= Start.")

type ProcMap []Entry

func (this ProcMap) ParseAppend(rdr io.Reader) (ProcMap, error) {
	if this == nil {
		this = make(ProcMap, 0, 16)
	}
	tok := bufio.NewScanner(rdr)
	for tok.Scan() {
		entry, err := ParseEntry(bytes.NewReader(tok.Bytes()))
		if err != nil {
			return nil, errs.Append(err, "Could not parse line entry")
		}
		this = append(this, entry)
	}
	if err := tok.Err(); err != nil {
		return nil, errs.Append(err, "Could not read entry lines")
	}
	return this, nil
}

type Entry struct {
	MemStart, MemEnd uint64
	Perms
	FileInfo
}

func (this Entry) Len() uint64 {
	return this.MemEnd - this.MemStart
}

func (this Entry) String() string {
	return fmt.Sprintf("%x-%x\t%s\t%x\t%x:%x\t%d\t%s",
		this.MemStart, this.MemEnd, this.Perms.String(), this.offset,
		this.devMajor, this.devMinor, this.inode, this.path)
}

func ParseEntry(rdr io.Reader) (entry Entry, err error) {
	var permStr string
	numParsed, err := fmt.Fscanf(rdr, "%x-%x %s %x %x:%x %d",
		&entry.MemStart, &entry.MemEnd, &permStr, &entry.offset,
		&entry.devMajor, &entry.devMinor, &entry.inode)
	if entry.MemEnd < entry.MemStart {
		return entry, ErrInvalidMemorySpan
	}

	if err != nil {
		switch numParsed {
		case 0:
			err = errs.Append(err, "Could not parse address range start")
		case 1:
			err = errs.Append(err, "Could not parse address range end")
		case 2:
			err = errs.Append(err, "Could not parse permissions")
		case 3:
			err = errs.Append(err, "Could not parse mapped file offset")
		case 4:
			err = errs.Append(err, "Could not parse mapped file major device number")
		case 5:
			err = errs.Append(err, "Could not parse mapped file minor device number")
		case 6:
			err = errs.Append(err, "Could not parse mapped file inode number")
		case 7:
			if entry.IsFileBacked() {
				err = errs.Append(err, "Could not parse mapped file path")
			} else {
				err = nil
			}
		default:
			err = errs.Append(err, "Could not parse entry, unexpected error")
		}
		if err != nil {
			return
		}
	}

	//Read the file path as that is all that remains -- TODO Cleanup!
	var filePathBuf bytes.Buffer
	var char = []byte{0}
	sawSlash := false
	for {
		if _, err = io.ReadFull(rdr, char); err != nil && err != io.EOF {
			return entry, errs.Append(err, "Could not read  file path")
		}
		if char[0] == '\n' || err == io.EOF {
			break
		}
		if !sawSlash && char[0] == '/' {
			sawSlash = true
		}
		if sawSlash || (char[0] != ' ' && char[0] != '\t') {
			filePathBuf.Write(char)
		}
	}
	entry.FileInfo.path = filePathBuf.String()
	// ** END Gross hack

	if entry.Perms, err = parsePerms(permStr); err != nil {
		err = errs.Append(err, "Could not parse permissions")
	}
	return
}

type Perms struct {
	read, write, exec, private bool
}

func (this Perms) String() string {
	var buf bytes.Buffer
	if this.read {
		buf.WriteByte('r')
	} else {
		buf.WriteByte('-')
	}
	if this.write {
		buf.WriteByte('w')
	} else {
		buf.WriteByte('-')
	}
	if this.exec {
		buf.WriteByte('x')
	} else {
		buf.WriteByte('-')
	}
	if this.private {
		buf.WriteByte('p')
	} else {
		buf.WriteByte('s')
	}
	return buf.String()
}

func (this Perms) Cvalue() int64 {
	const (
		PROT_NONE  = 0x000
		PROT_READ  = 0x001
		PROT_WRITE = 0x002
		PROT_EXEC  = 0x004
	)
	var cperms int64 = PROT_NONE
	if this.read {
		cperms |= PROT_READ
	}
	if this.write {
		cperms |= PROT_WRITE
	}
	if this.exec {
		cperms |= PROT_EXEC
	}
	return cperms
}

func parsePerms(permStr string) (perms Perms, err error) {
	if len(permStr) != 4 {
		return perms, errs.New("Line perms: %q, did not have four entries", permStr)
	}
	switch permStr[0] {
	case 'r':
		perms.read = true
	case '-':
	default:
		return perms, errs.New("Invalid value: %q, for read perm from perms: %q", permStr[0], permStr)
	}
	switch permStr[1] {
	case 'w':
		perms.write = true
	case '-':
	default:
		return perms, errs.New("Invalid value: %q, for write perm from perms: %q", permStr[1], permStr)
	}
	switch permStr[2] {
	case 'x':
		perms.exec = true
	case '-':
	default:
		return perms, errs.New("Invalid value: %q, for execute perm from perms: %q", permStr[2], permStr)
	}
	switch permStr[3] {
	case 'p':
		perms.private = true
	case 's':
	default:
		return perms, errs.New("Invalid value: %q, for private/shared perm from perms: %q", permStr[3], permStr)
	}
	return
}

type FileInfo struct {
	offset             uint64
	devMajor, devMinor uint32
	inode              uint64
	path               string
}

func (this FileInfo) IsFileBacked() bool {
	return this.devMajor != 0 && this.devMinor != 0 && this.inode != 0 &&
		len(this.path) > 0 && this.path[0] != '['
}

func (this FileInfo) Path() string {
	return this.path
}
