package pfiles

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lib/errs"
)

type FileEntry struct {
	FileHandle int
	Path       string
	Type       os.FileMode
	Pos, Flags int
}

func (this FileEntry) String() string {
	typeDesc := "regular file"
	switch {
	case this.Type.IsDir():
		typeDesc = "directory"
	case !this.Type.IsRegular():
		typeDesc = "special file"
		if strings.Contains(this.Path, "pty") || strings.Contains(this.Path, "pts") {
			typeDesc = "pseudo-terminal"
		} else if strings.Contains(this.Path, "tty") {
			typeDesc = "hardware-terminal (console)"
			if strings.Contains(this.Path, "ttyS") {
				typeDesc = "hardware-terminal (serial port)"
			}
		}
	}
	return fmt.Sprintf("File handle: %d, Path: %q, Type/Mode: %o (%s: %s), Seek position: %d, Flags: %d",
		this.FileHandle, this.Path, this.Type, typeDesc, this.Type, this.Pos, this.Flags)
}

func GetOpenFiles(PID int) ([]FileEntry, error) {
	fdEntries := newEntries()
	err := filepath.Walk(fmt.Sprintf("/proc/%d/fd/", PID), fdEntries.walk)
	return fdEntries.Entries(), err
}

type entries []FileEntry

func newEntries() entries {
	return entries(make([]FileEntry, 0, 4))
}

func (this entries) Entries() []FileEntry {
	return []FileEntry(this)
}

//Used for walking /proc/<PID>/{fd, fdinfo}, see
// http://man7.org/linux/man-pages/man5/proc.5.html for proc layout details
func (this *entries) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsNotExist(err) { //File have have been closed since this operation started
			return nil
		}
		return errs.Append(err, "Error walking file tree while on: %s", path)
	}

	if info.IsDir() { //We don't expet directories in /prox/<PID>/fd/ but we can ignore them
		return nil
	}
	fileHandle, err := strconv.Atoi(info.Name())
	if err != nil {
		return errs.Append(err, "Error: files in '/proc/<PID>/fd' are expected to have integer names, see: http://man7.org/linux/man-pages/man5/proc.5.html")
	}

	//Get target of symlink in: /proc/<PID>/fd/<symlink>
	targetPath, err := os.Readlink(path)
	if err == os.ErrNotExist {
		return nil
	} else if err != nil {
		return errs.Append(err, "Readlink error on: %s ", path)
	}

	//Stat target to find out what kind of file it is
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) { //File have have been closed since this operation started
			return nil
		}
		return errs.Append(err, "Could not stat symlink target: %s", targetPath)
	}

	//Get fileinfo file for each file descriptor from proc/self/fdinfo/<name>
	fdInfoPath := filepath.Join(strings.TrimSuffix(filepath.Dir(path), "fd"), "fdinfo", info.Name())
	pos, flags, err := readFileDescInfoFile(fdInfoPath)
	if err != nil {
		if os.IsNotExist(err) { //File have have been closed since this operation started
			return nil
		}
		return errs.Append(err, "Could read description of file handle for: %s from: %s", path, fdInfoPath)
	}

	//Append new entry on to ourself
	*this = append(*this, FileEntry{
		FileHandle: fileHandle,
		Path:       targetPath,
		Type:       targetInfo.Mode(),
		Pos:        pos,
		Flags:      flags,
	})
	return nil
}

func readFileDescInfoFile(path string) (pos int, flags int, err error) {
	fdInfoBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, 0, errs.Append(err, "Could not read file descriptor info in: %s", path)
	}
	if pos, flags, err = parseFileDescInfo(bytes.NewReader(fdInfoBytes)); err != nil {
		return 0, 0, errs.Append(err, "Could not parse file descriptor info file with contents: %q", fdInfoBytes)
	}
	return
}

func parseFileDescInfo(fdInfoStrm io.Reader) (pos int, flags int, err error) {
	//Get position (text decimal)
	if _, err = fmt.Fscanf(fdInfoStrm, "pos: %d\n", &pos); err != nil {
		errs.Append(err, `Could not extract file descriptor position ("pos: X")`)
	}
	//Get mode (text octal)
	if _, err = fmt.Fscanf(fdInfoStrm, "flags: %o\n", &flags); err != nil {
		errs.Append(err, `Could not extract file descriptor flags ("flags: X")`)
	}
	//there are other attributes like mnt_id that are interesting but may not exist
	// on old kernels and are not required
	return
}
