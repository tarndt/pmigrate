package iotimeout

import (
	"io"
	"time"
)

type readDeadliner interface {
	io.Reader
	SetReadDeadline(deadline time.Time) error
}

type writeDeadliner interface {
	io.Writer
	SetWriteDeadline(timeout time.Time) error
}

func WrapReadTimeout(rdr io.Reader, timeout time.Duration) io.Reader {
	if deadlineRdr, ok := rdr.(readDeadliner); ok {
		if wrapper, err := newDeadlineReader(deadlineRdr, timeout); err == nil {
			return wrapper
		}
	}
	return newTimeoutReader(rdr, timeout)
}

func WrapWriteTimeout(wtr io.Writer, timeout time.Duration) io.Writer {
	if deadlineWtr, ok := wtr.(writeDeadliner); ok {
		if wrapper, err := newDeadlineWriter(deadlineWtr, timeout); err == nil {
			return wrapper
		}
	}
	return newTimeoutWriter(wtr, timeout)
}
