//nolint: lll
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
	// Fixing URL tests
	{"gemini://gemini.circumlunar.space/%64%6f%63%73/%66%61%71%2e%67%6d%69", "gemini://gemini.circumlunar.space/docs/faq.gmi"},
	{"gemini://example.com/蛸", "gemini://example.com/%E8%9B%B8"},
	{"gemini://gemini.circumlunar.space/%64%6f%63%73/;;.'%66%61%71蛸%2e%67%6d%69", "gemini://gemini.circumlunar.space/docs/%3B%3B.%27faq%E8%9B%B8.gmi"},
	{"gemini://example.com/?%2Ch%64ello蛸", "gemini://example.com/?%2Chdello%E8%9B%B8"},
	// IPv6 tests, see #195
	{"gemini://[::1]", "gemini://[::1]/"},
	{"gemini://[::1]:1965", "gemini://[::1]/"},
	{"gemini://[::1]/test", "gemini://[::1]/test"},
	{"gemini://[::1]:1965/test", "gemini://[::1]/test"},
}

func TestNormalizeURL(t *testing.T) {
	for _, tt := range normalizeURLTests {
		actual := normalizeURL(tt.u)
		if actual != tt.expected {
			t.Errorf("normalizeURL(%s): expected %s, actual %s", tt.u, tt.expected, actual)
		}
	}
}
