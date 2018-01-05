package main

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"strings"

	"github.com/golang/snappy"

	"lib/errs"
	"lib/transpenc"
)

type NoopWtrCloser struct {
	io.Writer
}

func (this NoopWtrCloser) Close() error { return nil }

func getDestCompressor(dstWtr io.WriteCloser, compress string, transpEnc *transpenc.TranportEncoding) (io.WriteCloser, error) {
	compress = strings.ToLower(compress)
	transpEnc.CompressAlgo = compress

	switch compress {
	case "none", "":
		transpEnc.CompressAlgo = "none"
		return dstWtr, nil //noop
	case "gzip":
		return gzip.NewWriter(dstWtr), nil
	case "snappy":
		return NoopWtrCloser{snappy.NewWriter(dstWtr)}, nil
	case "flate":
		return flate.NewWriter(dstWtr, flate.DefaultCompression)
	}
	return nil, errs.New("Unknown compression method: %q, please use none, gzip, flate or snappy.", compress)
}
