package pmaps

import (
	"errors"
	"io"
	"os"

	"lib/errs"
)

var ErrUnreadable = errors.New("Entry is marked as unreadable.")

func ReadEntry(entry Entry, memFile *os.File, dst io.Writer, followFiles bool, backingFiles map[string]*os.File) (int, error, map[string]*os.File) {
	var buf []byte
	var err error
	var n int
	if entry.IsFileBacked() && followFiles {
		fin, isPresent := backingFiles[entry.path]
		if !isPresent {
			if fin, err = os.Open(entry.path); err != nil {
				return 0, errs.Append(err, "Could not open backing file: %s to read %X-%X", entry.path, entry.MemStart, entry.MemEnd), backingFiles
			}
			backingFiles[entry.path] = fin
		}
		buf = make([]byte, int(entry.Len()))
		n, err = fin.ReadAt(buf, int64(entry.offset))
	}
	if n < 1 {
		//Warning DRAGON: according to "man proc 5" aka. proc(5); this file can be
		// used to access the pages of a process's  memory through:
		// open(2), read(2), and lseek(2).
		// os.File.ReadAt uses the Pread (vs. read) system call and somtimes fails..)
		if _, err = memFile.Seek(int64(entry.MemStart), os.SEEK_SET); err != nil {
			return 0, errs.Append(err, "Could not seek to %X in memory file: %s", entry.MemStart, memFile.Name()), backingFiles
		}
		n, err := io.CopyN(dst, memFile, int64(entry.Len()))
		if err != nil {
			return 0, errs.Append(err, "Could not read %X-%X from memory file: %s", entry.MemStart, entry.MemEnd, memFile.Name()), backingFiles
		}
		return int(n), nil, backingFiles
	}
	if uint64(n) != entry.Len() {
		return 0, errs.Append(err, "Reading %X-%X from memory file: %s resulting in %d bytes being read when %d where expected", entry.MemStart, entry.MemEnd, memFile.Name(), n, entry.Len()), backingFiles
	}
	if _, err := dst.Write(buf); err != nil {
		return 0, errs.Append(err, "Could not write content read from memory file: %s, to provided destination writer", memFile.Name()), backingFiles
	}
	return n, nil, backingFiles
}
