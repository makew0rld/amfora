package subscriptions

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	urlPkg "net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

var (
	ErrSaving           = errors.New("couldn't save JSON to disk")
	ErrNotSuccess       = errors.New("status 20 not returned")
	ErrNotFeed          = errors.New("not a valid feed")
	ErrTooManyRedirects = errors.New("redirected more than 5 times")
)

var writeMu = sync.Mutex{} // Prevent concurrent writes to subscriptions.json file

// LastUpdated is the time when the in-memory data was last updated.
// It can be used to know if the subscriptions page should be regenerated.
var LastUpdated time.Time

// Init should be called after config.Init.
func Init() error {
	f, err := os.Open(config.SubscriptionPath)
	if err == nil {
		// File exists and could be opened

		fi, err := f.Stat()
		if err == nil && fi.Size() > 0 {
			// File is not empty

			jsonBytes, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return fmt.Errorf("read subscriptions.json error: %w", err)
			}
			err = json.Unmarshal(jsonBytes, &data)
			if err != nil {
				return fmt.Errorf("subscriptions.json is corrupted: %w", err)
			}
		}
		f.Close()
	} else if !os.IsNotExist(err) {
		// There's an error opening the file, but it's not bc is doesn't exist
		return fmt.Errorf("open subscriptions.json error: %w", err)
	}

	if data.Feeds == nil {
		data.Feeds = make(map[string]*gofeed.Feed)
	}
	if data.Pages == nil {
		data.Pages = make(map[string]*pageJSON)
	}

	LastUpdated = time.Now()

	if viper.GetInt("subscriptions.update_interval") > 0 {
		// Update subscriptions every so often
		go func() {
			for {
				updateAll()
				time.Sleep(time.Duration(viper.GetInt("subscriptions.update_interval")) * time.Second)
			}
		}()
	} else {
		// User disabled automatic updates
		// So just update once at the beginning
		go updateAll()
	}

	return nil
}

// IsSubscribed returns true if the URL is already subscribed to,
// whether a feed or page.
func IsSubscribed(url string) bool {
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
	if r == nil {
		return nil, false
	}

	// Check mediatype and filename
	if mediatype != "application/atom+xml" && mediatype != "application/rss+xml" && mediatype != "application/json+feed" &&
		filename != "atom.xml" && filename != "feed.xml" && filename != "feed.json" &&
		!strings.HasSuffix(filename, ".atom") && !strings.HasSuffix(filename, ".rss") &&
		!strings.HasSuffix(filename, ".xml") {
		// No part of the above is true
		return nil, false
	}
	feed, err := gofeed.NewParser().Parse(r)
	if feed == nil {
		return nil, false
	}
	return feed, err == nil
}

func writeJSON() error {
	writeMu.Lock()
	defer writeMu.Unlock()

	data.Lock()
	jsonBytes, err := json.MarshalIndent(&data, "", "  ")
	data.Unlock()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config.SubscriptionPath, jsonBytes, 0666)
	if err != nil {
		return err
	}

	return nil
}

// AddFeed stores a feed.
// It can be used to update a feed for a URL, although the package
// will handle that on its own.
func AddFeed(url string, feed *gofeed.Feed) error {
	if feed == nil {
		panic("feed is nil")
	}

	// Remove any unused fields to save memory and disk space
	feed.Image = nil
	feed.Generator = ""
	feed.Categories = nil
	feed.DublinCoreExt = nil
	feed.ITunesExt = nil
	feed.Custom = nil
	feed.Link = ""
	feed.Links = nil
	for _, item := range feed.Items {
		item.Description = ""
		item.Content = ""
		item.Image = nil
		item.Categories = nil
		item.Enclosures = nil
		item.DublinCoreExt = nil
		item.ITunesExt = nil
		item.Extensions = nil
		item.Custom = nil
		item.Link = "" // Links is used instead
	}

	data.feedMu.Lock()
	oldFeed, ok := data.Feeds[url]
	if !ok || !reflect.DeepEqual(feed, oldFeed) {
		// Feeds are different, or there was never an old one

		LastUpdated = time.Now()
		data.Feeds[url] = feed
		data.feedMu.Unlock()
		err := writeJSON()
		if err != nil {
			return ErrSaving
		}
	} else {
		data.feedMu.Unlock()
	}
	return nil
}

// AddPage stores a page to track for changes.
// It can be used to update the page as well, although the package
// will handle that on its own.
func AddPage(url string, r io.Reader) error {
	if r == nil {
		return nil
	}

	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return err
	}
	newHash := fmt.Sprintf("%x", h.Sum(nil))

	data.pageMu.Lock()
	_, ok := data.Pages[url]
	if !ok || data.Pages[url].Hash != newHash {
		// Page content is different, or it didn't exist

		LastUpdated = time.Now()
		data.Pages[url] = &pageJSON{
			Hash:    newHash,
			Changed: time.Now().UTC(),
		}

		data.pageMu.Unlock()
		err := writeJSON()
		if err != nil {
			return ErrSaving
		}
	} else {
		data.pageMu.Unlock()
	}

	return nil
}

// getResource returns a URL and Response for the given URL.
// It will follow up to 5 redirects, and if there is a permanent
// redirect it will return the new URL. Otherwise the URL will
// stay the same. THe returned URL will never be empty.
//
// If there is over 5 redirects the error will be ErrTooManyRedirects.
// ErrNotSuccess, as well as other fetch errors will also be returned.
func getResource(url string) (string, *gemini.Response, error) {
	res, err := client.Fetch(url)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return url, nil, err
	}

	if res.Status == gemini.StatusSuccess {
		// No redirects
		return url, res, nil
	}

	parsed, err := urlPkg.Parse(url)
	if err != nil {
		return url, nil, err
	}

	i := 0
	redirs := make([]int, 0)
	urls := make([]*urlPkg.URL, 0)

	// Loop through redirects
	for (res.Status == gemini.StatusRedirectPermanent || res.Status == gemini.StatusRedirectTemporary) && i < 5 {
		redirs = append(redirs, res.Status)
		urls = append(urls, parsed)

		tmp, err := parsed.Parse(res.Meta)
		if err != nil {
			// Redirect URL returned by the server is invalid
			return url, nil, err
		}
		parsed = tmp

		// Make the new request
		res, err := client.Fetch(parsed.String())
		if err != nil {
			if res != nil {
				res.Body.Close()
			}
			return url, nil, err
		}

		i++
	}

	// Two possible options here:
	// - Never redirected, got error on start
	// - No more redirects, other status code
	// - Too many redirects

	if i == 0 {
		// Never redirected or succeeded
		return url, res, ErrNotSuccess
	}

	if i < 5 {
		// The server stopped redirecting after <5 redirects

		if res.Status == gemini.StatusSuccess {
			// It ended by succeeding

			for j := range redirs {
				if redirs[j] == gemini.StatusRedirectTemporary {
					if j == 0 {
						// First redirect is temporary
						return url, res, nil
					}
					// There were permanent redirects before this one
					// Return the URL of the latest permanent redirect
					return urls[j-1].String(), res, nil
				}
			}
			// They were all permanent redirects
			return urls[len(urls)-1].String(), res, nil
		}

		// It stopped because there was a non-redirect, non-success response
		return url, res, ErrNotSuccess
	}

	// Too many redirects, return original
	return url, nil, ErrTooManyRedirects
}

func updateFeed(url string) {
	newURL, res, err := getResource(url)
	if err != nil {
		return
	}

	mediatype, _, err := mime.ParseMediaType(res.Meta)
	if err != nil {
		return
	}
	filename := path.Base(newURL)
	feed, ok := GetFeed(mediatype, filename, res.Body)
	if !ok {
		return
	}

	err = AddFeed(newURL, feed)
	if url != newURL && err == nil {
		// URL has changed, remove old one
		Remove(url) //nolint:errcheck
	}
}

func updatePage(url string) {
	newURL, res, err := getResource(url)
	if err != nil {
		return
	}

	err = AddPage(newURL, res.Body)
	if url != newURL && err == nil {
		// URL has changed, remove old one
		Remove(url) //nolint:errcheck
	}
}

// updateAll updates all subscriptions using workers.
// It only returns once all the workers are done.
func updateAll() {
	worker := func(jobs <-chan [2]string, wg *sync.WaitGroup) {
		// Each job is: [2]string{<type>, "url"}
		// where <type> is "feed" or "page"

		defer wg.Done()
		for j := range jobs {
			if j[0] == "feed" {
				updateFeed(j[1])
			} else if j[0] == "page" {
				updatePage(j[1])
			}
		}
	}

	var wg sync.WaitGroup

	data.RLock()
	numJobs := len(data.Feeds) + len(data.Pages)
	jobs := make(chan [2]string, numJobs)

	if numJobs == 0 {
		data.RUnlock()
		return
	}

	numWorkers := viper.GetInt("subscriptions.workers")
	if numWorkers < 1 {
		numWorkers = 1
	}

	// Start workers, waiting for jobs
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			worker(jobs, &wg)
		}()
	}

	// Get map keys in a slice

	feedKeys := make([]string, len(data.Feeds))
	i := 0
	for k := range data.Feeds {
		feedKeys[i] = k
		i++
	}

	pageKeys := make([]string, len(data.Pages))
	i = 0
	for k := range data.Pages {
		pageKeys[i] = k
		i++
	}
	data.RUnlock()

	for j := 0; j < numJobs; j++ {
		if j < len(feedKeys) {
			jobs <- [2]string{"feed", feedKeys[j]}
		} else {
			// In the Pages
			jobs <- [2]string{"page", pageKeys[j-len(feedKeys)]}
		}
	}
	close(jobs)

	wg.Wait()
}

// AllURLs returns all the subscribed-to URLS.
func AllURLS() []string {
	data.RLock()
	defer data.RUnlock()

	urls := make([]string, len(data.Feeds)+len(data.Pages))
	i := 0
	for k := range data.Feeds {
		urls[i] = k
		i++
	}
	for k := range data.Pages {
		urls[i] = k
		i++
	}

	return urls
}

// Remove removes a subscription from memory and from the disk.
// The URL must be provided. It will do nothing if the URL is
// not an actual subscription.
//
// It returns any errors that occurred when saving to disk.
func Remove(u string) error {
	data.Lock()
	// Just delete from both instead of using a loop to find it
	delete(data.Feeds, u)
	delete(data.Pages, u)
	data.Unlock()
	return writeJSON()
}
