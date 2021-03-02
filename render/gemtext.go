package render

import (
	"bufio"
	"io"
)

// Renderer for gemtext. Other Renderers are in renderer.go.

type GemtextRenderer struct {
	r *io.PipeReader
	w *io.PipeWriter

	// scanner is used to process line by line.
	scanner *bufio.Scanner

	// scanWriter is used to send data to the scanner, which reads out of the other
	// end of the pipe.
	scanWriter *io.PipeWriter

	// lineEnd holds the rest of line when the Read call cuts off the line being returned.
	lineEnd []byte

	links chan string

	// numLinks is the number of links that exist so far.
	numLinks int
	// width is the number of columns to wrap to.
	width int
	// proxied is whether the request is through the gemini:// scheme.
	proxied bool

	// pre indicates whether the renderer is currently in a preformatted block
	// or not.
	pre bool
}

// NewGemtextRenderer.
//
// width is the number of columns to wrap to.
//
// proxied is whether the request is through the gemini:// scheme.
// If it's not a gemini:// page, set this to true.
func NewGemtextRenderer(width int, proxied bool) *GemtextRenderer {
	pr, pw := io.Pipe()
	scanReader, scanWriter := io.Pipe()
	scanner := bufio.NewScanner(scanReader)
	links := make(chan string, 10)

	return &GemtextRenderer{
		r:          pr,
		w:          pw,
		scanner:    scanner,
		scanWriter: scanWriter,
		lineEnd:    make([]byte, 0),
		links:      links,
		numLinks:   0,
		width:      width,
		proxied:    proxied,
		pre:        false,
	}
}

func (r *GemtextRenderer) Links() <-chan string {
	return r.links
}

func (r *GemtextRenderer) Write(p []byte) (n int, err error) {
	// Just write to the scanner, all logic is in Read()
	return r.scanWriter.Write(p)
}
