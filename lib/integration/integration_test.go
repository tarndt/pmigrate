package integration

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"lib/preader"
	"lib/pwriter"
)

/* TestIntegration uses the countforever test program to:
1. Start process
2. Snapshot it
3. Let n = the last number wrote to stdout
4. Restore it from snapshot
5. Verify the next number written to stdout is n+1
*/
func TestIntegration(t *testing.T) {
	//Start the test countforever process
	countProg, stdout := startCountProg(t)
	defer countProg.Process.Kill()
	countCh := parseUintStrm(t, stdout)
	<-countCh //Read the first val to make sure child has executed

	//Construct the snapshot writer; write to memory
	captureBuf := new(bytes.Buffer)
	snapWtr := pwriter.NewProcSnapshotWriter(captureBuf)
	defer snapWtr.Close()

	//Start the process reader
	procRdr, err := preader.NewProcReader(countProg.Process)
	if err != nil {
		t.Fatalf("Could not attach to test process; Details:\n\t%s", err)
	}
	defer procRdr.Close()

	//Consume the process
	if err = snapWtr.Consume(procRdr); err != nil {
		t.Fatalf("Could not capture state of the test process", err)
	}
	//Get the last value the test process wrote to stdout
	targetLastVal := getLastValue(countCh)

	//Read the captured snapshot
	snapRdr, err := preader.NewProcSnapReader(bytes.NewReader(captureBuf.Bytes()))
	if err != nil {
		t.Fatalf("Could not read process state from source; Details:\n\t%s", err)
	}
	defer snapRdr.Close()

	//Restore the snapshot
	pipeOut, pipeIn := io.Pipe()
	restoredCountCh := parseUintStrm(t, pipeOut)
	iosinks := pwriter.DefaultStdioSinks()
	iosinks.Stdout = pipeIn
	procWriter := pwriter.NewProcWriterCustStdio("../../pthaw/pload/ploader", iosinks)
	go func() {
		if err := procWriter.Consume(snapRdr); err != nil {
			panic(err)
		}
	}()

	//Read first value from restored process and verify it
	restoredFirstVal := <-restoredCountCh
	if restoredFirstVal != targetLastVal+1 {
		t.Fatal("Test process's last value: %d, but the restored process's first value was: %d and not %d as expected.", targetLastVal, restoredFirstVal, targetLastVal+1)
	}
}

func startCountProg(t *testing.T) (*exec.Cmd, io.ReadCloser) {
	countProg := exec.Command("../../testprogs/countforever")
	stdout, err := countProg.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := countProg.Start(); err != nil {
		t.Fatal(err)
	}
	return countProg, stdout
}

func parseUintStrm(t *testing.T, instrm io.Reader) chan uint64 {
	countCh := make(chan uint64, 256)
	go func() {
		scnr, val := bufio.NewScanner(instrm), uint64(0)
		var err error
		for scnr.Scan() {
			if val, err = strconv.ParseUint(string(scnr.Bytes()), 10, 64); err != nil {
				t.Fatal(err)
				close(countCh)
				return
			}
			countCh <- val
		}
	}()
	return countCh
}

func getLastValue(valCh chan uint64) uint64 {
	const readTimeout = time.Millisecond * 10

	var val uint64
	var ok bool
	var lastValRead time.Time

	for {
		select {
		case val, ok = <-valCh:
			if !ok {
				return val
			}
			lastValRead = time.Now()
		default:
			if time.Now().After(lastValRead.Add(readTimeout)) {
				return val
			}
			time.Sleep(readTimeout)
		}
	}
}
