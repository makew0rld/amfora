// Package cache provides an interface for a cache of strings, aka text/gemini pages.
// It is fully thread safe.
package cache

import (
	"net/url"
	"sync"

	"github.com/makeworld-the-better-one/amfora/structs"
)

var pages = make(map[string]*structs.Page) // The actual cache
var urls = make([]string, 0)               // Duplicate of the keys in the `pages` map, but in order of being added
var maxPages = 0                           // Max allowed number of pages in cache
var maxSize = 0                            // Max allowed cache size in bytes
var lock = sync.RWMutex{}

// SetMaxPages sets the max number of pages the cache can hold.
// A value <= 0 means infinite pages.
func SetMaxPages(max int) {
	maxPages = max
}

// SetMaxSize sets the max size the cache can be, in bytes.
// A value <= 0 means infinite size.
func SetMaxSize(max int) {
	maxSize = max
}

func removeIndex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func removeUrl(url string) {
	for i := range urls {
		if urls[i] == url {
			urls = removeIndex(urls, i)
			return
		}
	}
}

// Add adds a page to the cache, removing earlier pages as needed
// to keep the cache inside its limits.
//
// If your page is larger than the max cache size, the provided page
// will silently not be added to the cache.
func Add(p *structs.Page) {
	if p.Url == "" {
		// Just in case, don't waste cache on new tab page
		return
	}
	// Never cache pages with query strings, to reduce unexpected behaviour
	parsed, err := url.Parse(p.Url)
	if err == nil && parsed.RawQuery != "" {
		return
	}

	if p.Size() > maxSize && maxSize > 0 {
		// This page can never be added
		return
	}

	// Remove earlier pages to make room for this one
	// There should only ever be 1 page to remove at most,
	// but this handles more just in case.
	for NumPages() >= maxPages && maxPages > 0 {
		Remove(urls[0])
	}
	// Do the same but for cache size
	for Size()+p.Size() > maxSize && maxSize > 0 {
		Remove(urls[0])
	}

	lock.Lock()
	defer lock.Unlock()
	pages[p.Url] = p
	// Remove the URL if it was already there, then add it to the end
	removeUrl(p.Url)
	urls = append(urls, p.Url)
}

// Remove will remove a page from the cache.
// Even if the page doesn't exist there will be no error.
func Remove(url string) {
	lock.Lock()
	defer lock.Unlock()
	delete(pages, url)
	removeUrl(url)
}

// Clear removes all pages from the cache.
func Clear() {
	lock.Lock()
	defer lock.Unlock()
	pages = make(map[string]*structs.Page)
	urls = make([]string, 0)
}

// Size returns the approx. current size of the cache in bytes.
func Size() int {
	lock.RLock()
	defer lock.RUnlock()
	n := 0
	for _, page := range pages {
		n += page.Size()
	}
	return n
}

func NumPages() int {
	lock.RLock()
	defer lock.RUnlock()
	return len(pages)
}

// Get returns the page struct, and a bool indicating if the page was in the cache or not.
// An empty page struct is returned if the page isn't in the cache
func Get(url string) (*structs.Page, bool) {
	lock.RLock()
	defer lock.RUnlock()
	p, ok := pages[url]
	return p, ok
}
