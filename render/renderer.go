package render

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"

	"code.rocketnine.space/tslocum/cview"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
)

// Renderer renderers network bytes into something that can be displayed on a
// cview.TextView.
//
// Calling Close when all writing is done is not a no-op, it will stop the the
// goroutine that runs for each Renderer, and will also allow the Links channel
// to be closed. Close should be called once all the data has been copied
//
// Write calls may block if the Lines channel buffer is full.
type Renderer interface {
	io.ReadWriteCloser

	// Links returns a channel that yields link URLs as they are parsed.
	// It is buffered. The channel will be closed when there won't be anymore links.
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
	readOut  *io.PipeReader
	readIn   *io.PipeWriter
	writeIn  *io.PipeWriter
	writeOut *io.PipeReader
}

func NewPlaintextRenderer() *PlaintextRenderer {
	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()
	ren := PlaintextRenderer{
		readOut:  pr1,
		readIn:   pw1,
		writeIn:  pw2,
		writeOut: pr2,
	}
	go ren.handler()
	return &ren
}

// handler is supposed to run in a goroutine as soon as the renderer is created.
// It handles the buffering and parsing in the background.
func (ren *PlaintextRenderer) handler() {
	scanner := bufio.NewScanner(ren.writeOut)
	scanner.Split(ScanLines)

	for scanner.Scan() {
		//nolint:errcheck
		ren.readIn.Write(cview.EscapeBytes(scanner.Bytes()))
	}
	if err := scanner.Err(); err != nil {
		// Close the ends this func touches, shouldn't matter really
		ren.writeOut.CloseWithError(err)
		ren.readIn.CloseWithError(err)
	}
}

func (ren *PlaintextRenderer) Write(p []byte) (n int, err error) {
	return ren.writeIn.Write(p)
}

func (ren *PlaintextRenderer) Read(p []byte) (n int, err error) {
	return ren.readOut.Read(p)
}

func (ren *PlaintextRenderer) Close() error {
	// Close user-facing ends of the pipes. Shouldn't matter which ends though
	ren.writeIn.Close()
	ren.readOut.Close()
	return nil
}

func (ren *PlaintextRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

// ANSIRenderer escapes text for cview usage, as well as converting ANSI codes
// into cview tags if the config allows it.
type ANSIRenderer struct {
	readOut    *io.PipeReader
	readIn     *io.PipeWriter
	writeIn    *io.PipeWriter
	writeOut   *io.PipeReader
	ansiWriter io.Writer     // cview.ANSIWriter
	buf        *bytes.Buffer // Where ansiWriter writes to
}

// Regex for identifying ANSI color codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func NewANSIRenderer() *ANSIRenderer {
	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()
	var ansiWriter io.Writer = nil // When ANSI is disabled
	var buf bytes.Buffer

	if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
		// ANSI enabled
		ansiWriter = cview.ANSIWriter(&buf)
	}
	ren := ANSIRenderer{
		readOut:    pr1,
		readIn:     pw1,
		writeIn:    pw2,
		writeOut:   pr2,
		ansiWriter: ansiWriter,
		buf:        &buf,
	}
	go ren.handler()
	return &ren
}

// handler is supposed to run in a goroutine as soon as the renderer is created.
// It handles the buffering and parsing in the background.
func (ren *ANSIRenderer) handler() {
	// Go through lines, render, and write each line

	scanner := bufio.NewScanner(ren.writeOut)
	scanner.Split(ScanLines)

	for scanner.Scan() {
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

		ren.readIn.Write(line) //nolint:errcheck
	}

	if err := scanner.Err(); err != nil {
		// Close the ends this func touches, shouldn't matter really
		ren.writeOut.CloseWithError(err)
		ren.readIn.CloseWithError(err)
	}
}

func (ren *ANSIRenderer) Write(p []byte) (n int, err error) {
	return ren.writeIn.Write(p)
}

func (ren *ANSIRenderer) Read(p []byte) (n int, err error) {
	return ren.readOut.Read(p)
}

func (ren *ANSIRenderer) Close() error {
	// Close user-facing ends of the pipes. Shouldn't matter which ends though
	ren.writeIn.Close()
	ren.readOut.Close()
	return nil
}

func (ren *ANSIRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}
