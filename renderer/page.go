package renderer

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"os"
	"strings"
	"time"

	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/ianaindex"
)

var ErrTooLarge = errors.New("page content would be too large")
var ErrTimedOut = errors.New("page download timed out")
var ErrCantDisplay = errors.New("invalid content for a page")
var ErrBadEncoding = errors.New("unsupported encoding")
var ErrBadMediatype = errors.New("displayable mediatype is not handled in the code, implementation error")

// isUTF8 returns true for charsets that are compatible with UTF-8 and don't need to be decoded.
func isUTF8(charset string) bool {
	utfCharsets := []string{"", "utf-8", "us-ascii"}
	for _, s := range utfCharsets {
		if charset == s || strings.ToLower(charset) == s {
			return true
		}
	}
	return false
}

// decodeMeta returns the output of mime.ParseMediaType, but handles the empty
// META which is equal to "text/gemini; charset=utf-8" according to the spec.
func decodeMeta(meta string) (string, map[string]string, error) {
	if meta == "" {
		return "text/gemini", make(map[string]string), nil
	}

	mediatype, params, err := mime.ParseMediaType(meta)

	if mediatype != "" && err != nil {
		// The mediatype was successfully decoded but there's some error with the params
		// Ignore the params
		return mediatype, make(map[string]string), nil
	}
	return mediatype, params, err
}

// CanDisplay returns true if the response is supported by Amfora
// for displaying on the screen.
// It also doubles as a function to detect whether something can be stored in a Page struct.
func CanDisplay(res *gemini.Response) bool {
	if gemini.SimplifyStatus(res.Status) != 20 {
		// No content
		return false
	}
	mediatype, params, err := decodeMeta(res.Meta)
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
func MakePage(url string, res *gemini.Response, width int, proxied bool) (*structs.Page, error) {
	if !CanDisplay(res) {
		return nil, ErrCantDisplay
	}

	buf := new(bytes.Buffer)
	_, err := io.CopyN(buf, res.Body, viper.GetInt64("a-general.page_max_size")+1)

	if err == nil {
		// Content was larger than max size
		return nil, ErrTooLarge
	} else if err != io.EOF {
		if os.IsTimeout(err) {
			// I would use
			// errors.Is(err, os.ErrDeadlineExceeded)
			// but that isn't supported before Go 1.15.

			return nil, ErrTimedOut
		}
		// Some other error
		return nil, err
	}
	// Otherwise, the error is EOF, which is what we want.

	mediatype, params, _ := decodeMeta(res.Meta)

	// Convert content first
	var utfText string
	if isUTF8(params["charset"]) {
		utfText = buf.String()
	} else {
		encoding, err := ianaindex.MIME.Encoding(params["charset"])
		if encoding == nil || err != nil {
			// Some encoding doesn't exist and wasn't caught in CanDisplay()
			return nil, ErrBadEncoding
		}
		utfText, err = encoding.NewDecoder().String(buf.String())
		if err != nil {
			return nil, err
		}
	}

	if mediatype == "text/gemini" {
		rendered, links := RenderGemini(utfText, width, proxied)
		return &structs.Page{
			Mediatype:    structs.TextGemini,
			RawMediatype: mediatype,
			URL:          url,
			Raw:          utfText,
			Content:      rendered,
			Links:        links,
			MadeAt:       time.Now(),
		}, nil
	} else if strings.HasPrefix(mediatype, "text/") {
		if mediatype == "text/x-ansi" || strings.HasSuffix(url, ".ans") || strings.HasSuffix(url, ".ansi") {
			// ANSI
			return &structs.Page{
				Mediatype:    structs.TextAnsi,
				RawMediatype: mediatype,
				URL:          url,
				Raw:          utfText,
				Content:      RenderANSI(utfText),
				Links:        []string{},
				MadeAt:       time.Now(),
			}, nil
		}

		// Treated as plaintext
		return &structs.Page{
			Mediatype:    structs.TextPlain,
			RawMediatype: mediatype,
			URL:          url,
			Raw:          utfText,
			Content:      RenderPlainText(utfText),
			Links:        []string{},
			MadeAt:       time.Now(),
		}, nil
	}

	return nil, ErrBadMediatype
}
