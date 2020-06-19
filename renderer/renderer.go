// Package renderer provides functions to convert various data into a cview primitive.
// Example objects include a Gemini response, and an error.
//
// Rendered lines always end with \r\n, in an effort to be Window compatible.
package renderer

import (
	"errors"
	"io/ioutil"
	"mime"
	urlPkg "net/url"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
	"golang.org/x/text/encoding/ianaindex"
)

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
	if mediatype != "text/gemini" && mediatype != "text/plain" {
		// Amfora doesn't support other filetypes
		return false
	}
	// Check if there is an encoding for the charset already
	if params["charset"] == "" {
		return true // Means UTF-8
	}
	_, err = ianaindex.MIME.Encoding(params["charset"]) // Lowercasing is done inside
	return err == nil
}

// convertRegularGemini converts non-preformatted blocks of text/gemini
// into a cview-compatible format.
// It also returns a slice of link URLs.
// numLinks is the number of links that exist so far.
//
// Since this only works on non-preformatted blocks, renderGemini
// should always be used instead.
func convertRegularGemini(s string, numLinks int) (string, []string) {
	links := make([]string, 0)
	lines := strings.Split(s, "\n")
	wrappedLines := make([]string, 0) // Final result

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \r\t\n")

		if strings.HasPrefix(lines[i], "#") && viper.GetBool("a-general.color") {

			// Headings
			if strings.HasPrefix(lines[i], "###") {
				lines[i] = "[fuchsia::b]" + lines[i] + "[-::-]"
			}
			if strings.HasPrefix(lines[i], "##") {
				lines[i] = "[lime::b]" + lines[i] + "[-::-]"
			}
			if strings.HasPrefix(lines[i], "#") {
				lines[i] = "[red::b]" + lines[i] + "[-::-]"
			}

			// Links
		} else if strings.HasPrefix(lines[i], "=>") && len([]rune(lines[i])) >= 3 {
			// Trim whitespace and separate link from link text

			lines[i] = strings.Trim(lines[i][2:], " \t") // Remove `=>` part too
			delim := strings.IndexAny(lines[i], " \t")   // Whitespace between link and link text

			var url string
			var linkText string
			if delim == -1 {
				// No link text
				url = lines[i]
				linkText = url
			} else {
				// There is link text
				url = lines[i][:delim]
				linkText = strings.Trim(lines[i][delim:], " \t")
			}

			if strings.TrimSpace(lines[i]) == "" || strings.TrimSpace(url) == "" {
				// Link was just whitespace, reset it and move on
				lines[i] = "=>"
				wrappedLines = append(wrappedLines, lines[i])
				continue
			}

			links = append(links, url)

			if viper.GetBool("a-general.color") {
				if pU, _ := urlPkg.Parse(url); pU.Scheme == "" || pU.Scheme == "gemini" {
					// A gemini link
					// Add the link text in blue (in a region), and a gray link number to the left of it
					lines[i] = `[silver::b][` + strconv.Itoa(numLinks+len(links)) + "[]" + "[-::-]  " +
						`[dodgerblue]["` + strconv.Itoa(numLinks+len(links)-1) + `"]` + linkText + `[""][-]`
				} else {
					// Not a gemini link, use purple instead
					lines[i] = `[silver::b][` + strconv.Itoa(numLinks+len(links)) + "[]" + "[-::-]  " +
						`[#8700d7]["` + strconv.Itoa(numLinks+len(links)-1) + `"]` + linkText + `[""][-]`
				}
			} else {
				// No colours allowed
				lines[i] = `[::b][` + strconv.Itoa(numLinks+len(links)) + "[]  " +
					`["` + strconv.Itoa(numLinks+len(links)-1) + `"]` + linkText + `[""][-]`
			}

			// Lists
		} else if strings.HasPrefix(lines[i], "* ") {
			if viper.GetBool("a-general.bullets") {
				lines[i] = " 🞄" + lines[i][1:]
			}
			// Optionally list lines could be colored here too, if color is enabled
		}

		// Final processing of each line: wrapping

		if strings.TrimSpace(lines[i]) == "" {
			// Just add empty line without processing
			wrappedLines = append(wrappedLines, "")
		} else {
			if (strings.HasPrefix(lines[i], "[silver::b]") && viper.GetBool("a-general.color")) || strings.HasPrefix(lines[i], "[::b]") {
				// It's a link line, so don't wrap it
				wrappedLines = append(wrappedLines, lines[i])
			} else if strings.HasPrefix(lines[i], ">") {
				// It's a quote line, add extra quote symbols to the start of each wrapped line

				// Remove beginning quote and maybe space
				lines[i] = strings.TrimPrefix(lines[i], ">")
				lines[i] = strings.TrimPrefix(lines[i], " ")

				temp := cview.WordWrap(lines[i], config.GetWrapWidth())
				for i := range temp {
					temp[i] = "> " + temp[i]
				}
				wrappedLines = append(wrappedLines, temp...)
			} else {
				wrappedLines = append(wrappedLines, cview.WordWrap(lines[i], config.GetWrapWidth())...)
			}
		}
	}

	return strings.Join(wrappedLines, "\r\n"), links
}

// renderGemini converts text/gemini into a cview displayable format.
// It also returns a slice of link URLs.
func RenderGemini(s string) (string, []string) {
	s = cview.Escape(s)
	if viper.GetBool("a-general.color") {
		s = cview.TranslateANSI(s)
	}
	lines := strings.Split(s, "\n")

	links := make([]string, 0)

	// Process and wrap non preformatted lines
	rendered := "" // Final result
	pre := false
	buf := "" // Block of regular or preformatted lines
	for i := range lines {
		if strings.HasPrefix(lines[i], "```") {
			if pre {
				// In a preformatted block, so add the text as is
				// Don't add the current line with backticks
				rendered += buf
			} else {
				// Not preformatted, regular text
				ren, lks := convertRegularGemini(buf, len(links))
				links = append(links, lks...)
				rendered += ren
			}
			buf = "" // Clear buffer for next block
			pre = !pre
			continue
		}
		// Lines always end with \r\n for Windows compatibility
		buf += strings.TrimSuffix(lines[i], "\r") + "\r\n"
	}
	// Gone through all the lines, but there still is likely a block in the buffer
	if pre {
		// File ended without closing the preformatted block
		rendered += buf
	} else {
		// Not preformatted, regular text
		// Same code as in the loop above
		ren, lks := convertRegularGemini(buf, len(links))
		links = append(links, lks...)
		rendered += ren
	}

	return rendered, links
}

func MakePage(url string, res *gemini.Response) (*structs.Page, error) {
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
	if params["charset"] == "" || strings.ToLower(params["charset"]) == "us-ascii" {
		utfText = string(rawText)
	} else {
		encoding, _ := ianaindex.MIME.Encoding(params["charset"])
		utfText, err = encoding.NewDecoder().String(string(rawText))
		if err != nil {
			return nil, err
		}
	}

	if mediatype == "text/plain" {
		return &structs.Page{
			Url:     url,
			Content: utfText,
			Links:   []string{}, // Plaintext has no links
		}, nil
	}
	if mediatype == "text/gemini" {
		rendered, links := RenderGemini(utfText)
		return &structs.Page{
			Url:     url,
			Content: rendered,
			Links:   links,
		}, nil
	}

	return nil, errors.New("displayable mediatype is not handled in the code, implementation error")
}
