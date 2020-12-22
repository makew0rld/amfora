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

// escapeMeta santizes a META string for use within a cview modal.
func escapeMeta(meta string) string {
	return cview.Escape(strings.ReplaceAll(meta, "\n", ""))
}

// isValidTab indicates whether the passed tab is still being used, even if it's not currently displayed.
func isValidTab(t *tab) bool {
	tempTabs := tabs
	for i := range tempTabs {
		if tempTabs[i] == t {
			return true
		}
	}
	return false
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
		parsed.Host = parsed.Hostname()
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
