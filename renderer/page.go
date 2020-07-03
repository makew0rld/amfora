package renderer

import (
	"errors"
	"io/ioutil"
	"mime"
	"strings"

	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/go-gemini"
	"golang.org/x/text/encoding/ianaindex"
)

// isUTF8 returns true for charsets that are compatible with UTF-8 and don't need to be decoded.
func isUTF8(charset string) bool {
	utfCharsets := []string{"", "utf-8", "us-ascii"}
	for i := range utfCharsets {
		if strings.ToLower(charset) == utfCharsets[i] {
			return true
		}
	}
	return false
}

// CanDisplay returns true if the response is supported by Amfora
// for displaying on the screen.
// It also doubles as a function to detect whether something can be stored in a Page struct.
func CanDisplay(res *gemini.Response) bool {
	if gemini.SimplifyStatus(res.Status) != 20 {
		// No content
		return false
	}
	mediatype, params, err := mime.ParseMediaType(res.Meta)
	if err != nil {
		return false
	}
	if !strings.HasPrefix(mediatype, "text/") {
		// Amfora doesn't support other filetypes
		return false
	}
	if isUTF8(params["charset"]) {
		return true
	}
	enc, err := ianaindex.MIME.Encoding(params["charset"]) // Lowercasing is done inside
	// Encoding sometimes returns nil, see #3 on this repo and golang/go#19421
	return err == nil && enc != nil
}

// MakePage creates a formatted, rendered Page from the given network response and params.
// You must set the Page.Width value yourself.
func MakePage(url string, res *gemini.Response, width, leftMargin int) (*structs.Page, error) {
	if !CanDisplay(res) {
		return nil, errors.New("not valid content for a Page")
	}

	rawText, err := ioutil.ReadAll(res.Body) // TODO: Don't use all memory on large pages
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	mediatype, params, _ := mime.ParseMediaType(res.Meta)

	// Convert content first
	var utfText string
	if isUTF8(params["charset"]) {
		utfText = string(rawText)
	} else {
		encoding, err := ianaindex.MIME.Encoding(params["charset"])
		if encoding == nil || err != nil {
			// Some encoding doesn't exist and wasn't caught in CanDisplay()
			return nil, errors.New("unsupported encoding")
		}
		utfText, err = encoding.NewDecoder().String(string(rawText))
		if err != nil {
			return nil, err
		}
	}

	if mediatype == "text/gemini" {
		rendered, links := RenderGemini(utfText, width, leftMargin)
		return &structs.Page{
			Url:     url,
			Raw:     utfText,
			Content: rendered,
			Links:   links,
		}, nil
	} else if strings.HasPrefix(mediatype, "text/") {
		// Treated as plaintext

		// Add left margin
		var shifted string
		lines := strings.Split(utfText, "\n")
		for i := range lines {
			shifted += strings.Repeat(" ", leftMargin) + lines[i] + "\n"
		}

		return &structs.Page{
			Url:     url,
			Raw:     utfText,
			Content: shifted,
			Links:   []string{},
		}, nil
	}

	return nil, errors.New("displayable mediatype is not handled in the code, implementation error")
}
