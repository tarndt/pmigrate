package pmaps

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tarndt/errs"
)

const selfMap = "/proc/self/maps"

func getSelfMapping() (ProcMap, error) {
	fin, err := os.Open(selfMap)
	if err != nil {
		return nil, errs.Append(err, "Could not open file containing this processes virtual memory mappings: %s", selfMap)
	}
	defer fin.Close()
	var mappings ProcMap
	if mappings, err = mappings.ParseAppend(fin); err != nil {
		return nil, errs.Append(err, "Could not parse file containing this processes virtual memory mappings: %s; Details: %s", selfMap)
	}
	return mappings, nil
}

func TestParseAppendSelf(t *testing.T) {
	if _, err := getSelfMapping(); err != nil {
		t.Fatal(err)
	}
}

func TestParseAppendSample(t *testing.T) {
	const in = `00400000-0040b000 r-xp 00000000 08:01 7077896                            /bin/cat
	0400000-0040b000 r-xp 00000000 08:01 7077896                            /bin/path with spaces
0060a000-0060b000 r--p 0000a000 08:01 7077896                            /bin/cat
0060b000-0060c000 rw-p 0000b000 08:01 7077896                            /bin/cat
01304000-01325000 rw-p 00000000 00:00 0                                  [heap]
7f3c1db69000-7f3c1e24b000 r--p 00000000 08:01 11803499                   /usr/lib/locale/locale-archive
7f3c1e24b000-7f3c1e406000 r-xp 00000000 08:01 1311699                    /lib/x86_64-linux-gnu/libc-2.19.so
7f3c1e406000-7f3c1e605000 ---p 001bb000 08:01 1311699                    /lib/x86_64-linux-gnu/libc-2.19.so
7f3c1e605000-7f3c1e609000 r--p 001ba000 08:01 1311699                    /lib/x86_64-linux-gnu/libc-2.19.so
7f3c1e609000-7f3c1e60b000 rw-p 001be000 08:01 1311699                    /lib/x86_64-linux-gnu/libc-2.19.so
7f3c1e60b000-7f3c1e610000 rw-p 00000000 00:00 0 
7f3c1e610000-7f3c1e633000 r-xp 00000000 08:01 1311696                    /lib/x86_64-linux-gnu/ld-2.19.so
7f3c1e816000-7f3c1e819000 rw-p 00000000 00:00 0 
7f3c1e830000-7f3c1e832000 rw-p 00000000 00:00 0 
7f3c1e832000-7f3c1e833000 r--p 00022000 08:01 1311696                    /lib/x86_64-linux-gnu/ld-2.19.so
7f3c1e833000-7f3c1e834000 rw-p 00023000 08:01 1311696                    /lib/x86_64-linux-gnu/ld-2.19.so
7f3c1e834000-7f3c1e835000 rw-p 00000000 00:00 0 
7ffcb8acd000-7ffcb8aee000 rw-p 00000000 00:00 0                          [stack]
7ffcb8bac000-7ffcb8bae000 r-xp 00000000 00:00 0                          [vdso]
ffffffffff600000-ffffffffff601000 r-xp 00000000 00:00 0                  [vsyscall]
`
	var mappings ProcMap
	var err error
	if mappings, err = mappings.ParseAppend(strings.NewReader(in)); err != nil {
		t.Fatalf("Could not parse process virtual memory mapping sample; Details: %s", err)
	}
	expectedEntries := []Entry{
		Entry{MemStart: 4194304, MemEnd: 4239360, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 8, devMinor: 1, inode: 7077896, path: "/bin/cat"}},
		Entry{MemStart: 4194304, MemEnd: 4239360, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 8, devMinor: 1, inode: 7077896, path: "/bin/path with spaces"}},
		Entry{MemStart: 6332416, MemEnd: 6336512, Perms: Perms{read: true, write: false, exec: false, private: true}, FileInfo: FileInfo{offset: 40960, devMajor: 8, devMinor: 1, inode: 7077896, path: "/bin/cat"}},
		Entry{MemStart: 6336512, MemEnd: 6340608, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 45056, devMajor: 8, devMinor: 1, inode: 7077896, path: "/bin/cat"}},
		Entry{MemStart: 19939328, MemEnd: 20074496, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: "[heap]"}},
		Entry{MemStart: 139896173268992, MemEnd: 139896180486144, Perms: Perms{read: true, write: false, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 8, devMinor: 1, inode: 11803499, path: "/usr/lib/locale/locale-archive"}},
		Entry{MemStart: 139896180486144, MemEnd: 139896182300672, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 8, devMinor: 1, inode: 1311699, path: "/lib/x86_64-linux-gnu/libc-2.19.so"}},
		Entry{MemStart: 139896182300672, MemEnd: 139896184393728, Perms: Perms{read: false, write: false, exec: false, private: true}, FileInfo: FileInfo{offset: 1814528, devMajor: 8, devMinor: 1, inode: 1311699, path: "/lib/x86_64-linux-gnu/libc-2.19.so"}},
		Entry{MemStart: 139896184393728, MemEnd: 139896184410112, Perms: Perms{read: true, write: false, exec: false, private: true}, FileInfo: FileInfo{offset: 1810432, devMajor: 8, devMinor: 1, inode: 1311699, path: "/lib/x86_64-linux-gnu/libc-2.19.so"}},
		Entry{MemStart: 139896184410112, MemEnd: 139896184418304, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 1826816, devMajor: 8, devMinor: 1, inode: 1311699, path: "/lib/x86_64-linux-gnu/libc-2.19.so"}},
		Entry{MemStart: 139896184418304, MemEnd: 139896184438784, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: ""}},
		Entry{MemStart: 139896184438784, MemEnd: 139896184582144, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 8, devMinor: 1, inode: 1311696, path: "/lib/x86_64-linux-gnu/ld-2.19.so"}},
		Entry{MemStart: 139896186560512, MemEnd: 139896186572800, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: ""}},
		Entry{MemStart: 139896186667008, MemEnd: 139896186675200, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: ""}},
		Entry{MemStart: 139896186675200, MemEnd: 139896186679296, Perms: Perms{read: true, write: false, exec: false, private: true}, FileInfo: FileInfo{offset: 139264, devMajor: 8, devMinor: 1, inode: 1311696, path: "/lib/x86_64-linux-gnu/ld-2.19.so"}},
		Entry{MemStart: 139896186679296, MemEnd: 139896186683392, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 143360, devMajor: 8, devMinor: 1, inode: 1311696, path: "/lib/x86_64-linux-gnu/ld-2.19.so"}},
		Entry{MemStart: 139896186683392, MemEnd: 139896186687488, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: ""}},
		Entry{MemStart: 140723406819328, MemEnd: 140723406954496, Perms: Perms{read: true, write: true, exec: false, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: "[stack]"}},
		Entry{MemStart: 140723407732736, MemEnd: 140723407740928, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: "[vdso]"}},
		Entry{MemStart: 18446744073699065856, MemEnd: 18446744073699069952, Perms: Perms{read: true, write: false, exec: true, private: true}, FileInfo: FileInfo{offset: 0, devMajor: 0, devMinor: 0, inode: 0, path: "[vsyscall]"}},
	}
	for i, expectedEntry := range expectedEntries {
		if i > len(mappings) || !reflect.DeepEqual(expectedEntry, mappings[i]) {
			line := strings.Split(in, "\n")[i]
			t.Fatalf("Actual results of parsing sample process virtual memory mapping: %d, did not match.\n\tInput: %s\n\tExpected result: %+v\n\tActual result: %+v", i, line, expectedEntry, mappings[i])
		}
	}
}
