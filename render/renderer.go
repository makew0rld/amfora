package render

import (
	"bytes"
	"fmt"
	"io"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// Renderer renderers network bytes into something that can be displayed on a
// cview.TextView.
type Renderer interface {
	io.ReadWriter

	// Links returns a channel that yields Link URLs as they are parsed.
	// It is buffered. The channel might be closed to indicate links are supported
	// for this renderer.
	Links() <-chan string
}

type PlaintextRenderer struct {
	*io.PipeReader
	w *io.PipeWriter
}

func NewPlaintextRenderer() *PlaintextRenderer {
	pr, pw := io.Pipe()
	return &PlaintextRenderer{pr, pw}
}

func (r *PlaintextRenderer) Write(p []byte) (n int, err error) {
	// TODO: The escaping will fail if the Write bytes end in the middle of a tag
	// How can this be avoided by users of this func?
	return r.w.Write(cview.EscapeBytes(p))
}

func (r *PlaintextRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

type ANSIRenderer struct {
	*io.PipeReader
	pw         *io.PipeWriter
	ansiWriter io.Writer // cview.ANSIWriter
	buf        bytes.Buffer
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

func (r *ANSIRenderer) Write(p []byte) (n int, err error) {
	if r.ansiWriter == nil {
		// ANSI disabled
		return r.pw.Write(ansiRegex.ReplaceAll(p, []byte{}))
	}
	// ANSI enabled

	r.buf.Reset()
	r.ansiWriter.Write(p) // Shouldn't error because everything it writes to are all bytes.Buffer
	return r.pw.Write(
		// The ANSIWriter injects tags like [-:-:-]
		// but this will reset the background to use the user's terminal color.
		// These tags need to be replaced with resets that use the theme color.
		bytes.ReplaceAll(
			r.buf.Bytes(),
			[]byte("[-:-:-]"),
			[]byte(fmt.Sprintf("[-:%s:-]", config.GetColorString("bg"))),
		),
	)
}

func (r *ANSIRenderer) Links() <-chan string {
	ch := make(chan string)
	close(ch)
	return ch
}
