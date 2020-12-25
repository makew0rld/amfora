package cache

import (
	"sync"
)

// Functions for caching emoji favicons.
// See gemini://mozz.us/files/rfc_gemini_favicon.gmi for details.

var favicons = make(map[string]string) // domain to emoji
var favMu = sync.RWMutex{}

var KnownNoFavicon = "no"

// AddFavicon will add an emoji to the cache under that host.
// It does not verify that the string passed is actually an emoji.
// You can pass KnownNoFavicon as the emoji when a host doesn't have a valid favicon.
func AddFavicon(host, emoji string) {
	favMu.Lock()
	favicons[host] = emoji
	favMu.Unlock()
}

// ClearFavicons removes all favicons from the cache
func ClearFavicons() {
	favMu.Lock()
	favicons = make(map[string]string)
	favMu.Unlock()
}

// GetFavicon returns the favicon string for the host.
// It returns an empty string if there is no favicon cached.
// It might also return KnownNoFavicon to indicate that that host does not have
// a favicon at all.
func GetFavicon(host string) string {
	favMu.RLock()
	defer favMu.RUnlock()
	return favicons[host]
}

func NumFavicons() int {
	favMu.RLock()
	defer favMu.RUnlock()
	return len(favicons)
}

func RemoveFavicon(host string) {
	favMu.Lock()
	delete(favicons, host)
	favMu.Unlock()
}
