package pmaps

import (
	"bytes"
	"os"
	"testing"
)

const selfData = "/proc/self/mem"

func TestReadEntrySelf(t *testing.T) {
	mappings, err := getSelfMapping()
	if err != nil {
		t.Fatalf("Could not parse file containing this processes virtual memory mappings: %s; Details: %s", selfMap, err)
	}
	memFile, err := os.Open(selfData)
	if err != nil {
		t.Fatalf("Could not open file containing this processes memory contents: %s; Details: %s", selfData, err)
	}
	defer memFile.Close()
	backingFileCache := make(map[string]*os.File, 13)

	//Read memory contents
	var sum uint64
	var memCopy bytes.Buffer
	var n int
	for _, entry := range mappings {
		size := entry.Len()
		n, err, backingFileCache = ReadEntry(entry, memFile, &memCopy, true, backingFileCache)
		switch {
		case err == ErrUnreadable:
			t.Logf("Range not readable; %s", entry.Perms)
		case err != nil:
			t.Fatalf("Could not read entry %q; Details:\n\t%s", entry, err)
		case n != int(size):
			t.Fatalf("entry.Size() was: %s, but ReadEntry(..) returned %d bytes.", size, n)
		}
		sum += size
	}
	if uint64(memCopy.Len()) != sum {
		t.Fatalf("%d bytes were read from: %s, but: %q, indicated %d bytes.", memCopy.Len(), selfData, selfMap, sum)
	}
}
