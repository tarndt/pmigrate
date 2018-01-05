package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/tarndt/pmigrate/lib"
	"github.com/tarndt/pmigrate/lib/iotimeout"
	"github.com/tarndt/pmigrate/lib/preader"
	"github.com/tarndt/pmigrate/lib/pwriter"
	"github.com/tarndt/pmigrate/lib/transpenc"
)

func main() {
	var (
		PID                       int
		dest, compress, encrypt   string
		dialTimeout, writeTimeout time.Duration
		halt, debug               bool
	)
	flag.IntVar(&PID, "pid", -1, "PID of process to be frozen")
	flag.StringVar(&dest, "dest", "stdout", "Output sink: stdout | tcp|udp:host:port | unix:socketpath | snapshot-filepath")
	flag.StringVar(&compress, "compress", "none", "Compression mode: none | gzip | flate | snappy")
	flag.StringVar(&encrypt, "encrypt", "none", "Encryption mode: none | AES-CFB|AES-CTR|AES-OFB:keypath")
	flag.DurationVar(&dialTimeout, "dial-timeout", 0, "Optional: Duration to wait for socket level connection to be established")
	flag.DurationVar(&writeTimeout, "write-timeout", 0, "Optional: Duration to wait transmitting data to an active stream before timing out")
	flag.BoolVar(&halt, "halt", false, "Halt the target process after state capture and transmission is complete")
	flag.BoolVar(&debug, "debug", false, "Debug: true | false, if enabled outgoing data will be displayed")
	flag.Parse()

	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "pfrez: This utility must be executed as root.")
		os.Exit(1)
	}
	if PID == -1 {
		fmt.Fprintln(os.Stderr, "pfrez: The PID of the target process to be captured was not provided.")
		flag.Usage()
		os.Exit(1)
	}

	var wtr lib.StateConsumer
	if debug {
		wtr = pwriter.NewDebugConsumer()
	} else {
		var transpEnc transpenc.TranportEncoding

		dstWriter, err := getDestWriter(dest, dialTimeout)
		if err != nil {
			log.Fatalf("Could not create process state destination; Details:\n\t%s", err)
		}
		defer dstWriter.Close()

		var timeoutWtr io.Writer
		if writeTimeout > 0 {
			timeoutWtr = iotimeout.WrapWriteTimeout(dstWriter, writeTimeout)
		} else {
			timeoutWtr = dstWriter
		}

		dstEncyptor, err := getDestEncryptor(timeoutWtr, encrypt, &transpEnc)
		if err != nil {
			log.Fatalf("Could not create process state encryptor; Details:\n\t%s", err)
		}
		defer dstEncyptor.Close()

		dstCompressor, err := getDestCompressor(dstEncyptor, compress, &transpEnc)
		if err != nil {
			log.Fatalf("Could not create process state compressor; Details:\n\t%s", err)
		}
		defer dstCompressor.Close()

		outStrm := bufio.NewWriter(dstCompressor)
		defer outStrm.Flush()

		if err := transpEnc.Write(dstWriter); err != nil {
			log.Fatalf("Could not write transport encoding; Details:\n\t%s", err)
		}
		wtr = pwriter.NewProcSnapshotWriter(outStrm)
	}
	defer wtr.Close()

	targetProcess, err := os.FindProcess(PID)
	if err != nil {
		log.Fatalf("Could not find target process with PID: %d; Details:\n\t%s", PID, err)
	}

	rdr, err := preader.NewProcReader(targetProcess)
	if err != nil {
		log.Fatalf("Could not attach to process with PID: %d; Details:\n\t%s", PID, err)
	}
	defer rdr.Close()

	if err = wtr.Consume(rdr); err != nil {
		log.Fatalf("Could not capture state of target process with PID: %d and invocation command: %q; Details:\n\t%s", PID, rdr.GetName(), err)
	}
	if debug {
		os.Stdout.WriteString(wtr.DebugInfo())
	}
	if halt {
		if err = rdr.GetProcess().Kill(); err != nil {
			log.Printf("Could not halt target process as requested; Details:\n\t%s", err)
		}
	}
}
