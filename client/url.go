package client

// Functions that transform and normalize URLs
// Originally used to be in display/util.go
// Moved here for #115, so URLs in the [auth] config section could be normalized

import (
	"net/url"
	"strings"

	"github.com/makeworld-the-better-one/go-gemini"
	"golang.org/x/text/unicode/norm"
)

// See doc for NormalizeURL
func normalizeURL(u string) (*url.URL, string) {
	u = norm.NFC.String(u)

	tmp, err := gemini.GetPunycodeURL(u)
	if err != nil {
		return nil, u
	}
	u = tmp
	parsed, _ := url.Parse(u)

	if parsed.Scheme == "" {
		// Always add scheme
		parsed.Scheme = "gemini"
	} else if parsed.Scheme != "gemini" {
		// Not a gemini URL, nothing to do
		return nil, u
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

	return parsed, ""
}

// NormalizeURL attempts to make URLs that are different strings
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
func NormalizeURL(u string) string {
	pu, s := normalizeURL(u)
	if pu != nil {
		// Could be normalized, return it
		return pu.String()
	}
	// Return the best URL available up to that point
	return s
}

// FixUserURL will take a user-typed URL and add a gemini scheme to it if
// necessary. It is not the same as normalizeURL, and that func should still
// be used, afterward.
//
// For example "example.com" will become "gemini://example.com", but
// "//example.com" will be left untouched.
func FixUserURL(u string) string {
	if !strings.HasPrefix(u, "//") && !strings.HasPrefix(u, "gemini://") && !strings.Contains(u, "://") {
		// Assume it's a Gemini URL
		u = "gemini://" + u
	}
	return u
}
