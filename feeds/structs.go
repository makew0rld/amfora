package feeds

import (
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

/*
Example JSON.
{
	"feeds": {
		"url1": <gofeed.Feed>,
		"url2": <gofeed.Feed>,
	},
	"pages": {
		"url1": {
			"hash": <hash>,
			"updated": <time>
		},
		"url2": {
			"hash": <hash>,
			"updated": <time>
		}
	}
}

"pages" are the pages tracked for changes that aren't feeds.
The hash used is SHA-256.
The time is in RFC 3339 format, preferably in the UTC timezone.
*/

// Decoded JSON
type jsonData struct {
	feedMu sync.RWMutex
	pageMu sync.RWMutex
	Feeds  map[string]*gofeed.Feed `json:"feeds,omitempty"`
	Pages  map[string]*pageJson    `json:"pages,omitempty"`
}

type pageJson struct {
	Hash    string    `json:"hash"`
	Updated time.Time `json:"updated"`
}

var data jsonData // Global instance of jsonData - loaded from JSON and used

// PageEntry is a single item on a feed page.
// It is used both for tracked feeds and pages.
type PageEntry struct {
	Author    string
	Title     string
	URL       string
	Published time.Time
}

// PageEntries is new-to-old list of Entry structs, used to create a feed page.
// It should always be assumed to be sorted when used in other packages.
type PageEntries struct {
	sync.RWMutex
	Entries []*PageEntry
}

// Implement sort.Interface

func (e *PageEntries) Len() int {
	e.RLock()
	defer e.RUnlock()
	return len(e.Entries)
}

func (e *PageEntries) Less(i, j int) bool {
	e.RLock()
	defer e.RUnlock()
	return e.Entries[i].Published.Before(e.Entries[j].Published)
}

func (e *PageEntries) Swap(i, j int) {
	e.Lock()
	e.Entries[i], e.Entries[j] = e.Entries[j], e.Entries[i]
	e.Unlock()
}
