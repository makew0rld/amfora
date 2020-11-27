package subscriptions

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

// TODO: Test for deadlocks and whether there should be more
// goroutines for file writing or other things.

var (
	ErrSaving     = errors.New("couldn't save JSON to disk")
	ErrNotSuccess = errors.New("status 20 not returned")
	ErrNotFeed    = errors.New("not a valid feed")
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
		defer f.Close()

		fi, err := f.Stat()
		if err == nil && fi.Size() > 0 {
			// File is not empty
			dec := json.NewDecoder(f)
			err = dec.Decode(&data)
			if err != nil && err != io.EOF {
				return fmt.Errorf("subscriptions.json is corrupted: %w", err) //nolint:goerr113
			}
		}
	} else if !os.IsNotExist(err) {
		// There's an error opening the file, but it's not bc is doesn't exist
		return fmt.Errorf("open subscriptions.json error: %w", err) //nolint:goerr113
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
	logger.Log.Println("subscriptions.IsSubscribed called")

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
	logger.Log.Println("subscriptions.GetFeed called")

	if r == nil {
		return nil, false
	}

	// Check mediatype and filename
	if mediatype != "application/atom+xml" && mediatype != "application/rss+xml" && mediatype != "application/json+feed" &&
		filename != "atom.xml" && filename != "feed.xml" && filename != "feed.json" &&
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

func writeJSON() error {
	logger.Log.Println("subscriptions.writeJSON called")

	writeMu.Lock()
	defer writeMu.Unlock()

	f, err := os.OpenFile(config.SubscriptionPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Println("subscriptions.writeJSON error", err)
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	data.Lock()
	logger.Log.Println("subscriptions.writeJSON acquired data lock")
	err = enc.Encode(&data)
	data.Unlock()

	if err != nil {
		logger.Log.Println("subscriptions.writeJSON error", err)
	}

	return err
}

// AddFeed stores a feed.
// It can be used to update a feed for a URL, although the package
// will handle that on its own.
func AddFeed(url string, feed *gofeed.Feed) error {
	logger.Log.Println("subscriptions.AddFeed called")

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
	logger.Log.Println("subscriptions.AddPage called")

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

func updateFeed(url string) error {
	logger.Log.Println("subscriptions.updateFeed called")

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
	logger.Log.Println("subscriptions.updatePage called")

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

	return AddPage(url, res.Body)
}

// updateAll updates all subscriptions using workers.
// It only returns once all the workers are done.
func updateAll() {
	logger.Log.Println("subscriptions.updateAll called")

	// TODO: Is two goroutines the right amount?

	worker := func(jobs <-chan [2]string, wg *sync.WaitGroup) {
		// Each job is: [2]string{<type>, "url"}
		// where <type> is "feed" or "page"

		defer wg.Done()
		for j := range jobs {
			if j[0] == "feed" {
				updateFeed(j[1]) //nolint:errcheck
			} else if j[0] == "page" {
				updatePage(j[1]) //nolint:errcheck
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
		go func(i int) {
			logger.Log.Println("started worker", i)
			worker(jobs, &wg)
			logger.Log.Println("ended worker", i)
		}(w)
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
