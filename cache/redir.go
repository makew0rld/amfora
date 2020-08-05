package cache

import "sync"

// Functions for caching redirects.

var redirUrls = make(map[string]string) // map original URL to redirect
var redirMu = sync.RWMutex{}

// AddRedir adds a original-to-redirect pair to the cache.
func AddRedir(og, redir string) {
	redirMu.Lock()
	defer redirMu.Unlock()

	for k, v := range redirUrls {
		if og == v {
			// The original URL param is the redirect URL for `k`.
			// This means there is a chain: k -> og -> redir
			// The chain should be removed
			redirUrls[k] = redir
		}
		if redir == k {
			// There's a loop
			// The newer version is preferred
			delete(redirUrls, k)
		}
	}
	redirUrls[og] = redir
}

// ClearRedirs removes all redirects from the cache.
func ClearRedirs() {
	redirMu.Lock()
	redirUrls = make(map[string]string)
	redirMu.Unlock()
}

// Redirect takes the provided URL and returns a redirected version, if a redirect
// exists for that URL in the cache.
// If one does not then the original URL is returned.
func Redirect(u string) string {
	redirMu.RLock()
	defer redirMu.RUnlock()

	// A single lookup is enough, because AddRedir
	// removes loops and chains.
	redir, ok := redirUrls[u]
	if ok {
		return redir
	}
	return u
}

func NumRedirs() int {
	redirMu.RLock()
	defer redirMu.RUnlock()
	return len(redirUrls)
}
