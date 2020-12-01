// Package renderer provides functions to convert various data into a cview primitive.
// Example objects include a Gemini response, and an error.
//
// Rendered lines always end with \r\n, in an effort to be Window compatible.
package renderer

import (
	"fmt"
	urlPkg "net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// Regex for identifying ANSI color codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// RenderANSI renders plain text pages containing ANSI codes.
// Practically, it is used for the text/x-ansi.
func RenderANSI(s string, leftMargin int) string {
	s = cview.Escape(s)
	if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
		s = cview.TranslateANSI(s)
	} else {
		s = ansiRegex.ReplaceAllString(s, "")
	}
	var shifted string
	lines := strings.Split(s, "\n")
	for i := range lines {
		shifted += strings.Repeat(" ", leftMargin) + lines[i] + "\n"
	}
	return shifted
}

// RenderPlainText should be used to format plain text pages.
func RenderPlainText(s string, leftMargin int) string {
	var shifted string
	lines := strings.Split(cview.Escape(s), "\n")
	for i := range lines {
		shifted += strings.Repeat(" ", leftMargin) +
			"[" + config.GetColorString("regular_text") + "]" + lines[i] + "[-]" +
			"\n"
	}
	return shifted
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

// convertRegularGemini converts non-preformatted blocks of text/gemini
// into a cview-compatible format.
// Since this only works on non-preformatted blocks, RenderGemini
// should always be used instead.
//
// It also returns a slice of link URLs.
// numLinks is the number of links that exist so far.
// width is the number of columns to wrap to.
//
//
// proxied is whether the request is through the gemini:// scheme.
// If it's not a gemini:// page, set this to true.
func convertRegularGemini(s string, numLinks, width int, proxied bool) (string, []string) {
	links := make([]string, 0)
	lines := strings.Split(s, "\n")
	wrappedLines := make([]string, 0) // Final result

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \r\t\n")

		if strings.HasPrefix(lines[i], "#") {
			// Headings
			var tag string
			if viper.GetBool("a-general.color") {
				if strings.HasPrefix(lines[i], "###") {
					tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_3"))
				} else if strings.HasPrefix(lines[i], "##") {
					tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_2"))
				} else if strings.HasPrefix(lines[i], "#") {
					tag = fmt.Sprintf("[%s::b]", config.GetColorString("hdg_1"))
				}
				wrappedLines = append(wrappedLines, wrapLine(lines[i], width, tag, "[-::-]", true)...)
			} else {
				// Just bold, no colors
				wrappedLines = append(wrappedLines, wrapLine(lines[i], width, "[::b]", "[-::-]", true)...)
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
				if viper.GetBool("a-general.show_link") {
					linkText += " (" + url + ")"
				}
			}

			if strings.TrimSpace(lines[i]) == "" || strings.TrimSpace(url) == "" {
				// Link was just whitespace, reset it and move on
				lines[i] = "=>"
				wrappedLines = append(wrappedLines, lines[i])
				continue
			}

			links = append(links, url)
			num := numLinks + len(links) // Visible link number, one-indexed

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
				if !proxied && err == nil &&
					(pU.Scheme == "" || pU.Scheme == "gemini" || pU.Scheme == "about") {
					// A gemini link
					// Add the link text in blue (in a region), and a gray link number to the left of it
					// Those are the default colors, anyway

					wrappedLink = wrapLine(linkText, width,
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

					wrappedLink = wrapLine(linkText, width,
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

				wrappedLink = wrapLine(linkText, width,
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
		} else if strings.HasPrefix(lines[i], "* ") {
			if viper.GetBool("a-general.bullets") {
				// Wrap list item, and indent wrapped lines past the bullet
				wrappedItem := wrapLine(lines[i][1:], width,
					fmt.Sprintf("    [%s]", config.GetColorString("list_text")),
					"[-]", false)
				// Add bullet
				wrappedItem[0] = fmt.Sprintf(" [%s]\u2022", config.GetColorString("list_text")) +
					wrappedItem[0] + "[-]"
				wrappedLines = append(wrappedLines, wrappedItem...)
			}
			// Optionally list lines could be colored here too, if color is enabled
		} else if strings.HasPrefix(lines[i], ">") {
			// It's a quote line, add extra quote symbols and italics to the start of each wrapped line

			// Remove beginning quote and maybe space
			lines[i] = strings.TrimPrefix(lines[i], ">")
			lines[i] = strings.TrimPrefix(lines[i], " ")
			wrappedLines = append(wrappedLines,
				wrapLine(lines[i], width, fmt.Sprintf("[%s::i]> ", config.GetColorString("quote_text")),
					"[-::-]", true)...,
			)

		} else if strings.TrimSpace(lines[i]) == "" {
			// Just add empty line without processing
			wrappedLines = append(wrappedLines, "")
		} else {
			// Regular line, just wrap it
			wrappedLines = append(wrappedLines, wrapLine(lines[i], width,
				fmt.Sprintf("[%s]", config.GetColorString("regular_text")),
				"[-]", true)...)
		}
	}

	return strings.Join(wrappedLines, "\r\n"), links
}

// RenderGemini converts text/gemini into a cview displayable format.
// It also returns a slice of link URLs.
//
// width is the number of columns to wrap to.
// leftMargin is the number of blank spaces to prepend to each line.
//
// proxied is whether the request is through the gemini:// scheme.
// If it's not a gemini:// page, set this to true.
func RenderGemini(s string, width, leftMargin int, proxied bool) (string, []string) {
	s = cview.Escape(s)

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

				// Support ANSI color codes in preformatted blocks - see #59
				if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
					buf = cview.TranslateANSI(buf)
				} else {
					buf = ansiRegex.ReplaceAllString(buf, "")
				}

				rendered += fmt.Sprintf("[%s]", config.GetColorString("preformatted_text")) +
					buf + "[-]"
			} else {
				// Not preformatted, regular text

				// ANSI not allowed in regular text - see #59
				buf = ansiRegex.ReplaceAllString(buf, "")

				ren, lks := convertRegularGemini(buf, len(links), width, proxied)
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
		// Same code as in the loop above

		if viper.GetBool("a-general.color") && viper.GetBool("a-general.ansi") {
			buf = cview.TranslateANSI(buf)
		} else {
			buf = ansiRegex.ReplaceAllString(buf, "")
		}
		rendered += fmt.Sprintf("[%s]", config.GetColorString("preformatted_text")) +
			buf + "[-]"
	} else {
		// Not preformatted, regular text
		// Same code as in the loop above

		buf = ansiRegex.ReplaceAllString(buf, "")

		ren, lks := convertRegularGemini(buf, len(links), width, proxied)
		links = append(links, lks...)
		rendered += ren
	}

	if leftMargin > 0 {
		renLines := strings.Split(rendered, "\n")
		for i := range renLines {
			renLines[i] = strings.Repeat(" ", leftMargin) + renLines[i]
		}
		return strings.Join(renLines, "\n"), links
	}

	return rendered, links
}
