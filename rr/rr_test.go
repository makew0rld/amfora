package rr

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var r1 *RestartReader

func reset() {
	r1 = NewRestartReader(ioutil.NopCloser(strings.NewReader("1234567890")))
}

func TestRead(t *testing.T) {
	reset()
	p := make([]byte, 1)
	n, err := r1.Read(p)
	assert.Equal(t, 1, n, "should read one byte")
	assert.Equal(t, nil, err, "should be no error")
	assert.Equal(t, []byte{'1'}, p, "should have read one byte, '1'")
}

func TestRestart(t *testing.T) {
	reset()
	p := make([]byte, 4)
	r1.Read(p)

	r1.Restart()
	p = make([]byte, 5)
	n, err := r1.Read(p)
	assert.Equal(t, []byte("12345"), p, "should read the first 5 bytes again")
	assert.Equal(t, 5, n, "should have read 4 bytes")
	assert.Equal(t, nil, err, "err should be nil")

	r1.Restart()
	p = make([]byte, 4)
	n, err = r1.Read(p)
	assert.Equal(t, []byte("1234"), p, "should read the first 4 bytes again")
	assert.Equal(t, 4, n, "should have read 4 bytes")
	assert.Equal(t, nil, err, "err should be nil")
}
