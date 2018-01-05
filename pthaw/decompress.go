package main

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"

	"lib/errs"
)

func getSrcDecompressor(srcRdr io.Reader, compress string) (io.Reader, error) {
	switch compress {
	case "none", "":
		return srcRdr, nil //noop
	case "gzip":
		return gzip.NewReader(srcRdr)
	case "snappy":
		return ioutil.NopCloser(snappy.NewReader(srcRdr)), nil
	case "flate":
		return flate.NewReader(srcRdr), nil
	}
	return nil, errs.New("Unknown compression method: %q, understood algorithms are: none, gzip, flate or snappy.", compress)
}
