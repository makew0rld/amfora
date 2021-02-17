package display

import (
	"errors"
	"net/url"
	"strings"

	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
	"golang.org/x/text/unicode/norm"
)

// This file contains funcs that are small, self-contained utilities.

// makeContentLayout returns a flex that contains the given TextView
// along with the current correct left margin, as well as a single empty
// line at the top, for a top margin.
func makeContentLayout(tv *cview.TextView) *cview.Flex {
	// Create horizontal flex with the left margin as an empty space
	horiz := cview.NewFlex()
	horiz.SetDirection(cview.FlexColumn)
	horiz.AddItem(nil, leftMargin(), 0, false)
	horiz.AddItem(tv, 0, 1, true)

	// Create a vertical flex with the other one and a top margin
	vert := cview.NewFlex()
	vert.SetDirection(cview.FlexRow)
	vert.AddItem(nil, 1, 0, false)
	vert.AddItem(horiz, 0, 1, true)

	return vert
}

// makeTabLabel takes a string and adds spacing to it, making it
// suitable for display as a tab label.
func makeTabLabel(s string) string {
	return " " + s + " "
}

// tabNumber gets the index of the tab in the tabs slice. It returns -1
// if the tab is not in that slice.
func tabNumber(t *tab) int {
	tempTabs := tabs
	for i := range tempTabs {
		if tempTabs[i] == t {
			return i
		}
	}
	return -1
}

// escapeMeta santizes a META string for use within a cview modal.
func escapeMeta(meta string) string {
	return cview.Escape(strings.ReplaceAll(meta, "\n", ""))
}

// isValidTab indicates whether the passed tab is still being used, even if it's not currently displayed.
func isValidTab(t *tab) bool {
	return tabNumber(t) != -1
}

func leftMargin() int {
	return int(float64(termW) * viper.GetFloat64("a-general.left_margin"))
}

func textWidth() int {
	if termW <= 0 {
		// This prevent a flash of 1-column text on startup, when the terminal
		// width hasn't been initialized.
		return viper.GetInt("a-general.max_width")
	}

	rightMargin := leftMargin()
	if leftMargin() > 10 {
		// 10 is the max right margin
		rightMargin = 10
	}

	max := termW - leftMargin() - rightMargin
	if max < viper.GetInt("a-general.max_width") {
		return max
	}
	return viper.GetInt("a-general.max_width")
}

// resolveRelLink returns an absolute link for the given absolute link and relative one.
// It also returns an error if it could not resolve the links, which should be displayed
// to the user.
func resolveRelLink(t *tab, prev, next string) (string, error) {
	if !t.hasContent() {
		return next, nil
	}

	prevParsed, _ := url.Parse(prev)
	nextParsed, err := url.Parse(next)
	if err != nil {
		return "", errors.New("link URL could not be parsed") //nolint:goerr113
	}
	return prevParsed.ResolveReference(nextParsed).String(), nil
}

// normalizeURL attempts to make URLs that are different strings
// but point to the same place all look the same.
//
// Example: gemini://gus.guru:1965/ and //gus.guru/.
// This function will take both output the same URL each time.
//
// It will also percent-encode invalid characters, and decode chars
// that don't need to be encoded. It will also apply Unicode NFC
// normalization.
//
// The string passed must already be confirmed to be a URL.
// Detection of a search string vs. a URL must happen elsewhere.
//
// It only works with absolute URLs.
func normalizeURL(u string) string {
	u = norm.NFC.String(u)

	tmp, err := gemini.GetPunycodeURL(u)
	if err != nil {
		return u
	}
	u = tmp
	parsed, _ := url.Parse(u)

	if parsed.Scheme == "" {
		// Always add scheme
		parsed.Scheme = "gemini"
	} else if parsed.Scheme != "gemini" {
		// Not a gemini URL, nothing to do
		return u
	}

	parsed.User = nil    // No passwords in Gemini
	parsed.Fragment = "" // No fragments either
	if parsed.Port() == "1965" {
		// Always remove default port
		hostname := parsed.Hostname()
		if strings.Contains(hostname, ":") {
			parsed.Host = "[" + parsed.Hostname() + "]"
		} else {
			parsed.Host = parsed.Hostname()
		}
	}

	// Add slash to the end of a URL with just a domain
	// gemini://example.com -> gemini://example.com/
	if parsed.Path == "" {
		parsed.Path = "/"
	} else {
		// Decode and re-encode path
		// This removes needless encoding, like that of ASCII chars
		// And encodes anything that wasn't but should've been
		parsed.RawPath = strings.ReplaceAll(url.PathEscape(parsed.Path), "%2F", "/")
	}

	// Do the same to the query string
	un, err := gemini.QueryUnescape(parsed.RawQuery)
	if err == nil {
		parsed.RawQuery = gemini.QueryEscape(un)
	}

	return parsed.String()
}

// fixUserURL will take a user-typed URL and add a gemini scheme to it if
// necessary. It is not the same as normalizeURL, and that func should still
// be used, afterward.
//
// For example "example.com" will become "gemini://example.com", but
// "//example.com" will be left untouched.
func fixUserURL(u string) string {
	if !strings.HasPrefix(u, "//") && !strings.HasPrefix(u, "gemini://") && !strings.Contains(u, "://") {
		// Assume it's a Gemini URL
		u = "gemini://" + u
	}
	return u
}
