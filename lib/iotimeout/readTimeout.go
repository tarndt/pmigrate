package iotimeout

import (
	"io"
	"time"

	"lib/errs"
)

type deadlineReader struct {
	rdr         readDeadliner
	readTimeout time.Duration
}

func newDeadlineReader(rdr readDeadliner, readTimeout time.Duration) (*deadlineReader, error) {
	this := &deadlineReader{
		rdr:         rdr,
		readTimeout: readTimeout,
	}
	//Test setting deadlines to make sure everything is working, these will be
	//over-set when Read and Write are next called.
	if err := this.rdr.SetReadDeadline(time.Now().Add(this.readTimeout)); err != nil {
		return nil, errs.Append(err, "Test invokation of 'SetReadDeadline' failed!")
	}
	return this, nil
}

func (this *deadlineReader) Read(buf []byte) (int, error) {
	this.rdr.SetReadDeadline(time.Now().Add(this.readTimeout))
	return this.rdr.Read(buf)
}

type timeoutReader struct {
	rdr         io.Reader
	readTimeout time.Duration
}

func newTimeoutReader(rdr io.Reader, readTimeout time.Duration) *timeoutReader {
	return &timeoutReader{
		rdr:         rdr,
		readTimeout: readTimeout,
	}
}

func (this *timeoutReader) Read(buf []byte) (n int, err error) {
	ch := make(chan struct{}, 1)
	go func() {
		n, err = this.rdr.Read(buf)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return
	case <-time.After(this.readTimeout):
		return 0, errs.New("Timeout while waiting %s for read", this.readTimeout)
	}
}
