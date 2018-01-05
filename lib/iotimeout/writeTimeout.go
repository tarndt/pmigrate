package iotimeout

import (
	"io"
	"time"

	"lib/errs"
)

type deadlineWriter struct {
	wtr          writeDeadliner
	writeTimeout time.Duration
}

func newDeadlineWriter(wtr writeDeadliner, writeTimeout time.Duration) (*deadlineWriter, error) {
	this := &deadlineWriter{
		wtr:          wtr,
		writeTimeout: writeTimeout,
	}
	//Test setting deadlines to make sure everything is working, these will be
	//over-set when Read and Write are next called.
	if err := this.wtr.SetWriteDeadline(time.Now().Add(this.writeTimeout)); err != nil {
		return nil, errs.Append(err, "Test invokation of 'SetWriteDeadline' failed!")
	}
	return this, nil
}

func (this *deadlineWriter) Write(buf []byte) (int, error) {
	this.wtr.SetWriteDeadline(time.Now().Add(this.writeTimeout))
	return this.wtr.Write(buf)
}

type timeoutWriter struct {
	wtr          io.Writer
	writeTimeout time.Duration
}

func newTimeoutWriter(wtr io.Writer, writeTimeout time.Duration) *timeoutWriter {
	return &timeoutWriter{
		wtr:          wtr,
		writeTimeout: writeTimeout,
	}
}

func (this *timeoutWriter) Write(buf []byte) (n int, err error) {
	ch := make(chan struct{}, 1)
	go func() {
		n, err = this.wtr.Write(buf)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return
	case <-time.After(this.writeTimeout):
		return 0, errs.New("Timeout while waiting %s for read", this.writeTimeout)
	}
}
