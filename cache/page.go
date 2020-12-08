// Package cache provides an interface for a cache of strings, aka text/gemini pages, and redirects.
// It is fully thread safe.
package cache

import (
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

// SetMaxSize sets the max size the page cache can be, in bytes.
// A value <= 0 means infinite size.
func SetMaxSize(max int) {
	maxSize = max
}

func removeIndex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func removeURL(url string) {
	for i := range urls {
		if urls[i] == url {
			urls = removeIndex(urls, i)
			return
		}
	}
}

// AddPage adds a page to the cache, removing earlier pages as needed
// to keep the cache inside its limits.
//
// If your page is larger than the max cache size, the provided page
// will silently not be added to the cache.
func AddPage(p *structs.Page) {
	if p.URL == "" {
		// Just in case, these pages shouldn't be cached
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
		RemovePage(urls[0])
	}
	// Do the same but for cache size
	for SizePages()+p.Size() > maxSize && maxSize > 0 {
		RemovePage(urls[0])
	}

	lock.Lock()
	defer lock.Unlock()
	pages[p.URL] = p
	// Remove the URL if it was already there, then add it to the end
	removeURL(p.URL)
	urls = append(urls, p.URL)
}

// RemovePage will remove a page from the cache.
// Even if the page doesn't exist there will be no error.
func RemovePage(url string) {
	lock.Lock()
	defer lock.Unlock()
	delete(pages, url)
	removeURL(url)
}

// ClearPages removes all pages from the cache.
func ClearPages() {
	lock.Lock()
	defer lock.Unlock()
	pages = make(map[string]*structs.Page)
	urls = make([]string, 0)
}

// SizePages returns the approx. current size of the cache in bytes.
func SizePages() int {
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

// GetPage returns the page struct, and a bool indicating if the page was in the cache or not.
// An empty page struct is returned if the page isn't in the cache.
func GetPage(url string) (*structs.Page, bool) {
	lock.RLock()
	defer lock.RUnlock()
	p, ok := pages[url]
	return p, ok
}
