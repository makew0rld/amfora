package feeds

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mmcdole/gofeed"
)

var (
	ErrSaving     = errors.New("couldn't save JSON to disk")
	ErrNotSuccess = errors.New("status 20 not returned")
	ErrNotFeed    = errors.New("not a valid feed")
)

var writeMu = sync.Mutex{}

// Init should be called after config.Init.
func Init() error {
	defer config.FeedJson.Close()

	dec := json.NewDecoder(config.FeedJson)
	err := dec.Decode(&data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("feeds json is corrupted: %v", err)
	}
	return nil

	// TODO: Start pulling all feeds in another thread
}

// IsTracked returns true if the feed/page URL is already being tracked.
func IsTracked(url string) bool {
	data.feedMu.RLock()
	for u := range data.Feeds {
		if url == u {
			data.feedMu.RUnlock()
			return true
		}
	}
	data.feedMu.RUnlock()
	data.pageMu.RLock()
	for u := range data.Pages {
		if url == u {
			data.pageMu.RUnlock()
			return true
		}
	}
	data.pageMu.RUnlock()
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
	if feed == nil {
		return nil, false
	}
	return feed, err == nil
}

func writeJson() error {
	writeMu.Lock()
	defer writeMu.Unlock()

	f, err := os.OpenFile(config.FeedPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	data.feedMu.Lock()
	data.pageMu.Lock()
	err = enc.Encode(&data)
	data.feedMu.Unlock()
	data.pageMu.Unlock()

	return err
}

// AddFeed stores a feed.
// It can be used to update a feed for a URL, although the package
// will handle that on its own.
func AddFeed(url string, feed *gofeed.Feed) error {
	if feed == nil {
		panic("feed is nil")
	}

	sort.Sort(feed)
	// Remove any content to save memory and disk space
	for _, item := range feed.Items {
		item.Content = ""
	}

	data.feedMu.Lock()
	data.Feeds[url] = feed
	data.feedMu.Unlock()

	err := writeJson()
	if err != nil {
		// Don't use in-memory if it couldn't be saved
		data.feedMu.Lock()
		delete(data.Feeds, url)
		data.feedMu.Unlock()
		return ErrSaving
	}
	return nil
}

// AddPage stores a page URL to track for changes.
// Do not use it to update a page, as it only resets the hash.
func AddPage(url string) error {
	data.pageMu.Lock()
	data.Pages[url] = &pageJson{} // No hash yet
	data.pageMu.Unlock()

	err := writeJson()
	if err != nil {
		// Don't use in-memory if it couldn't be saved
		data.pageMu.Lock()
		delete(data.Pages, url)
		data.pageMu.Unlock()
		return ErrSaving
	}
	return nil
}

func updateFeed(url string) error {
	res, err := client.Fetch(url)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return err
	}
	defer res.Body.Close()

	if res.Status != gemini.StatusSuccess {
		return ErrNotSuccess
	}
	mediatype, _, err := mime.ParseMediaType(res.Meta)
	if err != nil {
		return err
	}
	filename := path.Base(url)
	feed, ok := GetFeed(mediatype, filename, res.Body)
	if !ok {
		return ErrNotFeed
	}
	return AddFeed(url, feed)
}

func updatePage(url string) error {
	res, err := client.Fetch(url)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return err
	}
	defer res.Body.Close()

	if res.Status != gemini.StatusSuccess {
		return ErrNotSuccess
	}
	h := sha256.New()
	if _, err := io.Copy(h, res.Body); err != nil {
		return err
	}
	data.pageMu.Lock()
	data.Pages[url] = &pageJson{
		Hash:    fmt.Sprintf("%x", h.Sum(nil)),
		Updated: time.Now().UTC(),
	}
	data.pageMu.Unlock()

	err = writeJson()
	if err != nil {
		// Don't use in-memory if it couldn't be saved
		data.pageMu.Lock()
		delete(data.Pages, url)
		data.pageMu.Unlock()
		return err
	}

	return nil
}

// updateAll updates all feeds and pages.
// It should run in goroutine at a regular interval.
func updateAll() {

}
