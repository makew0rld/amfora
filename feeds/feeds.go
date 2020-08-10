package feeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/mmcdole/gofeed"
)

/*
Example JSON.
{
	"feeds": {
		"url1": <gofeed.Feed>,
		"url2: <gofeed.Feed>"
	},
	"pages": {
		"url1": "hash",
		"url2": "hash"
	}
}

"pages" are the pages tracked for changes that aren't feeds.
The hash is SHA-256.

*/

// Decoded JSON
type feedJson struct {
	Feeds map[string]*gofeed.Feed `json:"feeds"`
	Pages map[string]string       `json:"pages"`
}

var data feedJson

var ErrSaving = errors.New("couldn't save JSON to disk")

// Init should be called after config.Init.
func Init() error {
	defer config.FeedJson.Close()

	dec := json.NewDecoder(config.FeedJson)
	err := dec.Decode(&data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("feeds json is corrupted: %v", err)
	}
	return nil
}

// IsTracked returns true of the feed/page URL is already being tracked.
func IsTracked(url string) bool {
	for u := range data.Feeds {
		if url == u {
			return true
		}
	}
	for u := range data.Pages {
		if url == u {
			return true
		}
	}
	return false
}

// GetFeed returns a Feed object and a bool indicating whether the passed
// content was actually recognized as a feed.
func GetFeed(mediatype, filename string, r io.Reader) (*gofeed.Feed, bool) {
	// Check mediatype and filename
	if mediatype != "application/atom+xml" && mediatype != "application/rss+xml" &&
		filename != "atom.xml" && filename != "feed.xml" &&
		!strings.HasSuffix(filename, ".atom") && !strings.HasSuffix(filename, ".rss") {
		// No part of the above is true
		return nil, false
	}
	feed, err := gofeed.NewParser().Parse(r)
	return feed, err == nil
}

func writeJson() error {
	f, err := os.OpenFile(config.FeedPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err = enc.Encode(&data)
	return err
}

// AddFeed stores a feed.
func AddFeed(url string, feed *gofeed.Feed) error {
	sort.Sort(feed)
	data.Feeds[url] = feed
	err := writeJson()
	if err != nil {
		return ErrSaving
	}
	return nil
}

// AddPage stores a page URL to track for changes.
func AddPage(url string) error {
	data.Pages[url] = "" // No hash yet
	err := writeJson()
	if err != nil {
		return ErrSaving
	}
	return nil
}
