package display

import (
	"testing"
)

var normalizeURLTests = []struct {
	u        string
	expected string
}{
	{"gemini://example.com:1965/", "gemini://example.com/"},
	{"gemini://example.com", "gemini://example.com/"},
	{"//example.com", "gemini://example.com/"},
	{"//example.com:1965", "gemini://example.com/"},
	{"//example.com:123/", "gemini://example.com:123/"},
	{"gemini://example.com/", "gemini://example.com/"},
	{"gemini://example.com/#fragment", "gemini://example.com/"},
	{"gemini://example.com#fragment", "gemini://example.com/"},
	{"gemini://user@example.com/", "gemini://example.com/"},
	// Other schemes, URL isn't modified
	{"mailto:example@example.com", "mailto:example@example.com"},
	{"magnet:?xt=urn:btih:test", "magnet:?xt=urn:btih:test"},
	{"https://example.com", "https://example.com"},
}

func TestNormalizeURL(t *testing.T) {
	for _, tt := range normalizeURLTests {
		actual := normalizeURL(tt.u)
		if actual != tt.expected {
			t.Errorf("normalizeURL(%s): expected %s, actual %s", tt.u, tt.expected, actual)
		}
	}
}
