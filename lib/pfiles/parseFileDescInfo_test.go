package pfiles

import (
	"strings"
	"testing"
)

func TestParseFileDescInfo(t *testing.T) {
	type testCase struct {
		fileContents string
		expected     FileEntry
		shouldError  bool
	}

	testCases := []testCase{
		//0. Empty
		testCase{fileContents: "", shouldError: true},
		//1. Simple case
		testCase{fileContents: "pos:	0\nflags:	02004002", expected: FileEntry{Pos: 0, Flags: 02004002}},
		//2. Case with superfluous info
		testCase{fileContents: "pos:	0\nflags:	02004000\ninotify wd:4 ino:7c1f sdev:3 mask:800afce ignored_mask:0 fhandle-bytes:8 fhandle-type:1 f_handle:1f7c000000000000", expected: FileEntry{Pos: 0, Flags: 02004000}},
		//3. Case with non-zero position and zero flag
		testCase{fileContents: "pos:	13\nflags:	0", expected: FileEntry{Pos: 13, Flags: 0}},
		//4. Case with missing "pos"
		testCase{fileContents: "foobar: 2\nflags:	0123", shouldError: true},
		//5. Case with missing "flag"
		testCase{fileContents: "pos:	1\n", shouldError: true},
	}

	for i, testCase := range testCases {
		pos, flags, err := parseFileDescInfo(strings.NewReader(testCase.fileContents))
		if err != nil {
			if testCase.shouldError {
				continue
			}
			t.Fatalf("Test case: %d; Unexpected error: %q for input: %q", i, err, testCase.fileContents)
		} else if testCase.shouldError {
			t.Fatalf("Test case: %d; Unexpected success, parse was expected to fail (result had; pos: %d, flags: %o for input: %q)", i, pos, flags, testCase.fileContents)
		}
		if testCase.expected.Pos != pos {
			t.Fatalf("Test case: %d; Incorrect position found! Expected: %d, Actual: %d for input: %q", i, testCase.expected.Pos, pos, testCase.fileContents)
		} else if testCase.expected.Flags != flags {
			t.Fatalf("Test case: %d; Incorrect flag found! Expected: %o, Actual: %o for input: %q", i, testCase.expected.Flags, flags, testCase.fileContents)
		}
	}
}
