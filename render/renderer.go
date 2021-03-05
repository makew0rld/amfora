package render

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// Renderer renderers network bytes into something that can be displayed on a
// cview.TextView.
//
// Write calls may block if the Lines channel buffer is full.
//
// Current implementations don't actually implement io.Writer, and calling Write
// will panic. ReadFrom should be used instead.
type Renderer interface {
	io.ReadWriter
	io.ReaderFrom

	// Links returns a channel that yields Link URLs as they are parsed.
	// It is buffered. The channel might be closed to indicate links aren't supported
	// for this renderer.
	Links() <-chan string
}

// ScanLines is copied from bufio.ScanLines and is used with bufio.Scanner.
// The only difference is that this func doesn't get rid of the end-of-line marker.
// This is so that the number of read bytes can be counted correctly in ReadFrom.
//
// It also simplifes code by no longer having to append a newline character.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// PlaintextRenderer escapes text for cview usage and does nothing else.
type PlaintextRenderer struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func NewPlaintextRenderer() *PlaintextRenderer {
	pr, pw := io.Pipe()
	return &PlaintextRenderer{pr, pw}
}

func (ren *PlaintextRenderer) ReadFrom(r io.Reader) (int64, error) {
	// Go through lines and escape bytes and write each line
	// TODO: Should writes be buffered?

	var n int64
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLines)

	for scanner.Scan() {
		n += int64(len(scanner.Bytes()))

		//nolint:errcheck
		ren.w.Write(cview.EscapeBytes(scanner.Bytes()))
	}
	return n, scanner.Err()
}

// Write will panic, use ReadFrom instead.
func (ren *PlaintextRenderer) Write(p []byte) (n int, err error) {
	// This function would normally use cview.EscapeBytes
	// But the escaping will fail if the Write bytes end in the middle of a tag
	// So instead it just panics, because it should never be used.
	panic("func Write not allowed for PlaintextRenderer")
}

func (ren *PlaintextRenderer) Read(p []byte) (n int, err error) {
	return ren.r.Read(p)
}

func (ren *PlaintextRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

// ANSIRenderer escapes text for cview usage, as well as converting ANSI codes
// into cview tags if the config allows it.
type ANSIRenderer struct {
	r          *io.PipeReader
	w          *io.PipeWriter
	ansiWriter io.Writer    // cview.ANSIWriter
	buf        bytes.Buffer // Where ansiWriter writes to
}

func NewANSIRenderer() *ANSIRenderer {
	pr, pw := io.Pipe()

	var ansiWriter io.Writer = nil // When ANSI is disabled
	var buf bytes.Buffer

	if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
		// ANSI enabled
		ansiWriter = cview.ANSIWriter(&buf)
	}
	return &ANSIRenderer{pr, pw, ansiWriter, buf}
}

// Write will panic, use ReadFrom instead.
func (ren *ANSIRenderer) Write(p []byte) (n int, err error) {
	// This function would normally use cview.EscapeBytes among other things.
	// But the escaping will fail if the Write bytes end in the middle of a tag
	// So instead it just panics, because it should never be used.
	panic("func Write not allowed for ANSIRenderer")
}

func (ren *ANSIRenderer) ReadFrom(r io.Reader) (int64, error) {
	// Go through lines, render, and write each line
	// TODO: Should writes be buffered?

	var n int64
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLines)

	for scanner.Scan() {
		n += int64(len(scanner.Bytes()))
		line := scanner.Bytes()
		line = cview.EscapeBytes(line)

		if ren.ansiWriter == nil {
			// ANSI disabled
			line = ansiRegex.ReplaceAll(scanner.Bytes(), nil)
		} else {
			// ANSI enabled

			ren.buf.Reset()

			// Shouldn't error because everything it writes to are all bytes.Buffer
			ren.ansiWriter.Write(line) //nolint:errcheck

			// The ANSIWriter injects tags like [-:-:-]
			// but this will reset the background to use the user's terminal color.
			// These tags need to be replaced with resets that use the theme color.
			line = bytes.ReplaceAll(
				ren.buf.Bytes(),
				[]byte("[-:-:-]"),
				[]byte(fmt.Sprintf("[-:%s:-]", config.GetColorString("bg"))),
			)
		}

		ren.w.Write(line) //nolint:errcheck
	}

	return n, scanner.Err()
}

func (ren *ANSIRenderer) Read(p []byte) (n int, err error) {
	return ren.r.Read(p)
}

func (ren *ANSIRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}
