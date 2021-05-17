package render

import (
	"bufio"
	"fmt"
	"io"
	urlPkg "net/url"
	"strconv"
	"strings"

	"code.rocketnine.space/tslocum/cview"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
)

// Renderer for gemtext. Other Renderers are in renderer.go.

type GemtextRenderer struct {
	// Buffers and I/O

	readOut  *io.PipeReader
	readIn   *io.PipeWriter
	writeIn  *io.PipeWriter
	writeOut *io.PipeReader
	links    chan string

	// Configurable options

	// width is the number of columns to wrap to.
	width int
	// proxied is whether the request is through the gemini:// scheme.
	proxied      bool
	ansiEnabled  bool
	colorEnabled bool

	// State

	// pre indicates whether the renderer is currently in a preformatted block
	// or not.
	pre bool
	// numLinks is the number of links that exist so far.
	numLinks int
}

// wrapLine wraps a line to the provided width, and adds the provided prefix and suffix to each wrapped line.
// It recovers from wrapping panics and should never cause a panic.
// It returns a slice of lines, without newlines at the end.
//
// Set includeFirst to true if the prefix and suffix should be applied to the first wrapped line as well
func wrapLine(line string, width int, prefix, suffix string, includeFirst bool) []string {
	// Anonymous function to allow recovery from potential WordWrap panic
	var ret []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Use unwrapped line instead
				if includeFirst {
					ret = []string{prefix + line + suffix}
				} else {
					ret = []string{line}
				}
			}
		}()

		wrapped := cview.WordWrap(line, width)
		for i := range wrapped {
			if !includeFirst && i == 0 {
				continue
			}
			wrapped[i] = prefix + wrapped[i] + suffix
		}
		ret = wrapped
	}()
	return ret
}

// NewGemtextRenderer.
//
// width is the number of columns to wrap to.
//
// proxied is whether the request is through the gemini:// scheme.
// If it's not a gemini:// page, set this to true.
func NewGemtextRenderer(width int, proxied bool) *GemtextRenderer {
	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()

	ansiEnabled := false
	if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
		ansiEnabled = true
	}
	colorEnabled := false
	if viper.GetBool("a-general.color") {
		colorEnabled = true
	}

	ren := GemtextRenderer{
		readOut:      pr1,
		readIn:       pw1,
		writeIn:      pw2,
		writeOut:     pr2,
		links:        make(chan string, 10),
		width:        width,
		proxied:      proxied,
		ansiEnabled:  ansiEnabled,
		colorEnabled: colorEnabled,
	}
	go ren.handler()
	return &ren
}

// handler is supposed to run in a goroutine as soon as the renderer is created.
// It handles the buffering and parsing in the background.
func (ren *GemtextRenderer) handler() {
	// Go through lines, render, and write each line

	// Splits on lines and drops terminators, unlike the other renderers
	scanner := bufio.NewScanner(ren.writeOut)

	for scanner.Scan() {
		line := scanner.Text()

		// Process the one possibly invisible line
		if strings.HasPrefix(line, "```") {
			ren.pre = !ren.pre
			continue
		}

		// Render line and write it

		//nolint:errcheck
		ren.readIn.Write([]byte(ren.renderLine(line)))

	}

	// Everything has been read, no more links
	close(ren.links)

	if err := scanner.Err(); err != nil {
		// Close the ends this func touches, shouldn't matter really
		ren.writeOut.CloseWithError(err)
		ren.readIn.CloseWithError(err)
	}
}

// renderLine handles all lines except preformatted markings. The input line
// should not end with any line delimiters, but the output line does.
func (ren *GemtextRenderer) renderLine(line string) string {
	if ren.pre {
		if ren.ansiEnabled {
			line = cview.TranslateANSI(line)
			// The TranslateANSI function injects tags like [-:-:-]
			// but this will reset the background to use the user's terminal color.
			// These tags need to be replaced with resets that use the theme color.
			line = strings.ReplaceAll(line, "[-:-:-]",
				fmt.Sprintf("[%s:%s:-]", config.GetColorString("preformatted_text"), config.GetColorString("bg")),
			)

			// Set color at beginning and end of line to prevent background glitches
			// where the terminal background color slips through. This only happens on
			// preformatted blocks with ANSI characters.
			line = fmt.Sprintf("[%s]", config.GetColorString("preformatted_text")) +
				line + fmt.Sprintf("[%s:%s:-]", config.GetColorString("regular_text"), config.GetColorString("bg"))

		} else {
			line = ansiRegex.ReplaceAllString(line, "")
		}

		return line + "\n"
	}
	// Not preformatted, regular lines

	wrappedLines := make([]string, 0) // Final result

	// ANSI not allowed in regular text - see #59
	line = ansiRegex.ReplaceAllString(line, "")

	if strings.HasPrefix(line, "#") {
		// Headings
		var tag string
		if viper.GetBool("a-general.color") {
			if strings.HasPrefix(line, "###") {
				tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_3"))
			} else if strings.HasPrefix(line, "##") {
				tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_2"))
			} else if strings.HasPrefix(line, "#") {
				tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_1"))
			}
			wrappedLines = append(wrappedLines, wrapLine(line, ren.width, tag, "[-::-]", true)...)
		} else {
			// Just bold, no colors
			wrappedLines = append(wrappedLines, wrapLine(line, ren.width, "[::b]", "[-::-]", true)...)
		}

		// Links
	} else if strings.HasPrefix(line, "=>") && len([]rune(line)) >= 3 {
		// Trim whitespace and separate link from link text

		line = strings.Trim(line[2:], " \t")   // Remove `=>` part too
		delim := strings.IndexAny(line, " \t") // Whitespace between link and link text

		var url string
		var linkText string
		if delim == -1 {
			// No link text
			url = line
			linkText = url
		} else {
			// There is link text
			url = line[:delim]
			linkText = strings.Trim(line[delim:], " \t")
			if viper.GetBool("a-general.show_link") {
				linkText += " (" + url + ")"
			}
		}

		if strings.TrimSpace(line) == "" || strings.TrimSpace(url) == "" {
			// Link was just whitespace, return it
			return "=>\n"
		}

		ren.links <- url
		ren.numLinks++
		num := ren.numLinks // Visible link number, one-indexed

		var indent int
		if num > 99 {
			// Indent link text by 3 or more spaces
			indent = len(strconv.Itoa(num)) + 4 // +4 indent for spaces and brackets
		} else {
			// One digit and two digit links have the same spacing - see #60
			indent = 5 // +4 indent for spaces and brackets, and 1 for link number
		}

		// Spacing after link number: 1 or 2 spaces?
		var spacing string
		if num > 9 {
			// One space to keep it in line with other links - see #60
			spacing = " "
		} else {
			// One digit numbers use two spaces
			spacing = "  "
		}

		// Wrap and add link text
		// Wrap the link text, but add some spaces to indent the wrapped lines past the link number
		// Set the style tags
		// Add them to the first line

		var wrappedLink []string

		if viper.GetBool("a-general.color") {
			pU, err := urlPkg.Parse(url)
			if !ren.proxied && err == nil &&
				(pU.Scheme == "" || pU.Scheme == "gemini" || pU.Scheme == "about") {
				// A gemini link
				// Add the link text in blue (in a region), and a gray link number to the left of it
				// Those are the default colors, anyway

				wrappedLink = wrapLine(linkText, ren.width,
					strings.Repeat(" ", indent)+
						`["`+strconv.Itoa(num-1)+`"][`+config.GetColorString("amfora_link")+`]`,
					`[-][""]`,
					false, // Don't indent the first line, it's the one with link number
				)

				// Add special stuff to first line, like the link number
				wrappedLink[0] = fmt.Sprintf(`[%s::b][`, config.GetColorString("link_number")) +
					strconv.Itoa(num) + "[]" + "[-::-]" + spacing +
					`["` + strconv.Itoa(num-1) + `"][` + config.GetColorString("amfora_link") + `]` +
					wrappedLink[0] + `[-][""]`
			} else {
				// Not a gemini link

				wrappedLink = wrapLine(linkText, ren.width,
					strings.Repeat(" ", indent)+
						`["`+strconv.Itoa(num-1)+`"][`+config.GetColorString("foreign_link")+`]`,
					`[-][""]`,
					false, // Don't indent the first line, it's the one with link number
				)

				wrappedLink[0] = fmt.Sprintf(`[%s::b][`, config.GetColorString("link_number")) +
					strconv.Itoa(num) + "[]" + "[-::-]" + spacing +
					`["` + strconv.Itoa(num-1) + `"][` + config.GetColorString("foreign_link") + `]` +
					wrappedLink[0] + `[-][""]`
			}
		} else {
			// No colors allowed

			wrappedLink = wrapLine(linkText, ren.width,
				strings.Repeat(" ", len(strconv.Itoa(num))+4)+ // +4 for spaces and brackets
					`["`+strconv.Itoa(num-1)+`"]`,
				`[""]`,
				false, // Don't indent the first line, it's the one with link number
			)

			wrappedLink[0] = `[::b][` + strconv.Itoa(num) + "[][::-]  " +
				`["` + strconv.Itoa(num-1) + `"]` +
				wrappedLink[0] + `[""]`
		}

		wrappedLines = append(wrappedLines, wrappedLink...)

		// Lists
	} else if strings.HasPrefix(line, "* ") {
		if viper.GetBool("a-general.bullets") {
			// Wrap list item, and indent wrapped lines past the bullet
			wrappedItem := wrapLine(line[1:], ren.width,
				fmt.Sprintf("    [%s]", config.GetColorString("list_text")),
				"[-]", false)
			// Add bullet
			wrappedItem[0] = fmt.Sprintf(" [%s]\u2022", config.GetColorString("list_text")) +
				wrappedItem[0] + "[-]"
			wrappedLines = append(wrappedLines, wrappedItem...)
		} else {
			wrappedItem := wrapLine(line[1:], ren.width,
				fmt.Sprintf("    [%s]", config.GetColorString("list_text")),
				"[-]", false)
			// Add "*"
			wrappedItem[0] = fmt.Sprintf(" [%s]*", config.GetColorString("list_text")) +
				wrappedItem[0] + "[-]"
			wrappedLines = append(wrappedLines, wrappedItem...)

		}
		// Optionally list lines could be colored here too, if color is enabled
	} else if strings.HasPrefix(line, ">") {
		// It's a quote line, add extra quote symbols and italics to the start of each wrapped line

		if len(line) == 1 {
			// Just an empty quote line
			wrappedLines = append(wrappedLines, fmt.Sprintf("[%s::i]>[-::-]", config.GetColorString("quote_text")))
		} else {
			// Remove beginning quote and maybe space
			line = strings.TrimPrefix(line, ">")
			line = strings.TrimPrefix(line, " ")
			wrappedLines = append(wrappedLines,
				wrapLine(line, ren.width, fmt.Sprintf("[%s::i]> ", config.GetColorString("quote_text")),
					"[-::-]", true)...,
			)
		}

	} else if strings.TrimSpace(line) == "" {
		// Just add empty line without processing
		wrappedLines = append(wrappedLines, "")
	} else {
		// Regular line, just wrap it
		wrappedLines = append(wrappedLines, wrapLine(line, ren.width,
			fmt.Sprintf("[%s]", config.GetColorString("regular_text")),
			"[-]", true)...)
	}

	return strings.Join(wrappedLines, "\n") + "\n"
}

func (ren *GemtextRenderer) Write(p []byte) (n int, err error) {
	return ren.writeIn.Write(p)
}

func (ren *GemtextRenderer) Read(p []byte) (n int, err error) {
	return ren.readOut.Read(p)
}

func (ren *GemtextRenderer) Close() error {
	// Close user-facing ends of the pipes. Shouldn't matter which ends though
	ren.writeIn.Close()
	ren.readOut.Close()
	return nil
}

func (ren *GemtextRenderer) Links() <-chan string {
	return ren.links
}
