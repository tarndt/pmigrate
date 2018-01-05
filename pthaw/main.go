package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"lib/iotimeout"
	"lib/preader"
	"lib/pwriter"
	"lib/transpenc"
)

func main() {
	var (
		src, loaderPath, keyDir string
		readTimeout             time.Duration
		debug                   bool
	)

	runtime.LockOSThread() //This is needed to ensure PTRACE syscall interdiction always comes back the thread which is expecting the PTRACE events

	flag.StringVar(&src, "src", "stdin", "Input source: stdin | tcp|udp:port | unix:socketpath | snapshot-filepath")
	flag.StringVar(&loaderPath, "loader", "", "Optional: Alternate path to loader executable")
	flag.StringVar(&keyDir, "keydir", "", "Optional: Directory containing decryption keys")
	flag.DurationVar(&readTimeout, "read-timeout", 0, "Optional: Duration to wait for incomming data on an active stream before timing out")
	flag.BoolVar(&debug, "debug", false, "Debug: true | false, if enabled incomming data will be displayed")
	flag.Parse()

	if loaderPath == "" {
		loaderPath = filepath.Join(mustGetExecDir(), "ploader")
	}
	if keyDir == "" {
		keyDir = filepath.Join(mustGetExecDir(), "/")
	}

	srcRdr, err := getSourceReader(src)
	if err != nil {
		log.Fatalf("Could not open process state destination; Details:\n\t%s", err)
	}
	defer srcRdr.Close()

	var timeoutRdr io.Reader
	if readTimeout > 0 {
		timeoutRdr = iotimeout.WrapReadTimeout(srcRdr, readTimeout)
	} else {
		timeoutRdr = srcRdr
	}

	inStrm := bufio.NewReader(timeoutRdr)
	var transpEnc transpenc.TranportEncoding
	if err = transpenc.ReadTranportEncoding(inStrm, &transpEnc); err != nil {
		log.Fatalf("Could not read transport encoding of source stream; Details:\n\t%s", err)
	}

	srcDecryptor, err := getSrcDecyptor(inStrm, transpEnc.EncParams, keyDir)
	if err != nil {
		log.Fatalf("Could not source decryptor; Details:\n\t%s", err)
	}

	srcDecompressor, err := getSrcDecompressor(srcDecryptor, transpEnc.CompressAlgo)
	if err != nil {
		log.Fatalf("Could not source decompressor; Details:\n\t%s", err)
	}

	snapshotRdr, err := preader.NewProcSnapReader(bufio.NewReader(srcDecompressor))
	if err != nil {
		log.Fatalf("Could not read process state from source; Details:\n\t%s", err)
	}
	defer snapshotRdr.Close()

	if debug {
		debugWtr := pwriter.NewDebugConsumer()
		if err := debugWtr.Consume(snapshotRdr); err != nil {
			log.Fatalf("Could not consume process snapshot; Details:\n\t%s", err)
		}
		os.Stdout.WriteString(debugWtr.DebugInfo())
	} else {
		procWriter := pwriter.NewProcWriter(loaderPath)
		if err := procWriter.Consume(snapshotRdr); err != nil {
			log.Fatalf("Could not consume process snapshot; Details:\n\t%s", err)
		}
	}
}

var execDir string

func mustGetExecDir() string {
	if execDir != "" {
		return execDir
	}
	var err error
	if execDir, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		log.Fatalf("Could not determine this executable's location; Details:\n\t%s", err)
	}
	return execDir
}
