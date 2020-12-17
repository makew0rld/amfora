package rr

import (
	"errors"
	"io"
)

var ErrClosed = errors.New("RestartReader: closed")

type RestartReader struct {
	r   io.ReadCloser
	buf []byte

	// Where in the buffer we are. If it's equal to len(buf) then the reader
	// should be used.
	i int64
}

func (rr *RestartReader) Read(p []byte) (n int, err error) {
	if rr.buf == nil {
		return 0, ErrClosed
	}

	if rr.i >= int64(len(rr.buf)) {
		// Read new data
		tmp := make([]byte, len(p))
		n, err = rr.r.Read(tmp)
		if n > 0 {
			rr.buf = append(rr.buf, tmp[:n]...)
			copy(p, tmp[:n])
		}
		rr.i = int64(len(rr.buf))
		return
	}

	// Reading from buffer

	bufSize := len(rr.buf[rr.i:])

	if len(p) > bufSize {
		// It wants more data then what's in the buffer
		tmp := make([]byte, len(p)-bufSize)
		n, err = rr.r.Read(tmp)
		if n > 0 {
			rr.buf = append(rr.buf, tmp[:n]...)
		}
		copy(p, rr.buf[rr.i:])
		n += bufSize
		rr.i = int64(len(rr.buf))
		return
	}
	// All the required data is in the buffer
	end := rr.i + int64(len(p))
	copy(p, rr.buf[rr.i:end])
	rr.i = end
	n = len(p)
	err = nil
	return
}

// Restart causes subsequent Read calls to read from the beginning, instead
// of where they left off.
func (rr *RestartReader) Restart() {
	rr.i = 0
}

// Close clears the buffer and closes the underlying io.ReadCloser, returning
// its error.
func (rr *RestartReader) Close() error {
	rr.buf = nil
	return rr.r.Close()
}

// NewRestartReader creates and initializes a new RestartReader that reads from
// the provided io.ReadCloser.
func NewRestartReader(r io.ReadCloser) *RestartReader {
	return &RestartReader{
		r:   r,
		buf: make([]byte, 0),
	}
}
