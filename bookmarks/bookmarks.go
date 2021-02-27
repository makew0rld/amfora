package bookmarks

import (
	"encoding/base32"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
)

func Init() error {
	f, err := os.Open(config.BkmkPath)
	if err == nil {
		// File exists and could be opened

		fi, err := f.Stat()
		if err == nil && fi.Size() > 0 {
			// File is not empty

			xbelBytes, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return fmt.Errorf("read bookmarks.xml error: %w", err)
			}
			err = xml.Unmarshal(xbelBytes, &data)
			if err != nil {
				return fmt.Errorf("bookmarks.xml is corrupted: %w", err)
			}
		}
		f.Close()
	} else if !os.IsNotExist(err) {
		// There's an error opening the file, but it's not bc is doesn't exist
		return fmt.Errorf("open bookmarks.xml error: %w", err)
	}

	if data.Bookmarks == nil {
		data.Bookmarks = make([]*xbelBookmark, 0)
		data.Version = xbelVersion
	}

	if config.BkmkStore != nil {
		// There's still bookmarks stored in the old format
		// Add them and delete the file

		names, urls := oldBookmarks()
		for i := range names {
			data.Bookmarks = append(data.Bookmarks, &xbelBookmark{
				URL:  urls[i],
				Name: names[i],
			})
		}

		err := writeXbel()
		if err != nil {
			return fmt.Errorf("error saving old bookmarks into new format: %w", err)
		}

		err = os.Remove(config.OldBkmkPath)
		if err != nil {
			return fmt.Errorf(
				"couldn't delete old bookmarks file (%s), you must delete it yourself to prevent duplicate bookmarks: %w",
				config.OldBkmkPath,
				err,
			)
		}
		config.BkmkStore = nil
	}

	return nil
}

// oldBookmarks returns a slice of names and a slice of URLs of the
// bookmarks in config.BkmkStore.
func oldBookmarks() ([]string, []string) {
	bkmksMap, ok := config.BkmkStore.AllSettings()["bookmarks"].(map[string]interface{})
	if !ok {
		// No bookmarks stored yet, return empty map
		return []string{}, []string{}
	}

	names := make([]string, 0, len(bkmksMap))
	urls := make([]string, 0, len(bkmksMap))

	for b32Url, name := range bkmksMap {
		if n, ok := name.(string); n == "" || !ok {
			// name is not a string, or it's empty - ignore
			// Likely means it is a removed bookmark
			continue
		}
		url, err := base32.StdEncoding.DecodeString(strings.ToUpper(b32Url))
		if err != nil {
			// This would only happen if a user messed around with the bookmarks file
			continue
		}
		names = append(names, name.(string))
		urls = append(urls, string(url))
	}
	return names, urls
}

func writeXbel() error {
	xbelBytes, err := xml.MarshalIndent(&data, "", "    ")
	if err != nil {
		return err
	}

	xbelBytes = append(xbelHeader, xbelBytes...)
	err = ioutil.WriteFile(config.BkmkPath, xbelBytes, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Change the name of the bookmark at the provided URL.
func Change(url, name string) {
	for _, bkmk := range data.Bookmarks {
		if bkmk.URL == url {
			bkmk.Name = name
			writeXbel() //nolint:errcheck
			return
		}
	}
}

// Add will add a new bookmark.
func Add(url, name string) {
	data.Bookmarks = append(data.Bookmarks, &xbelBookmark{
		URL:  url,
		Name: name,
	})
	writeXbel() //nolint:errcheck
}

// Get returns the NAME of the bookmark, given the URL.
// It also returns a bool indicating whether it exists.
func Get(url string) (string, bool) {
	for _, bkmk := range data.Bookmarks {
		if bkmk.URL == url {
			return bkmk.Name, true
		}
	}
	return "", false
}

func Remove(url string) {
	for i, bkmk := range data.Bookmarks {
		if bkmk.URL == url {
			data.Bookmarks[i] = data.Bookmarks[len(data.Bookmarks)-1]
			data.Bookmarks = data.Bookmarks[:len(data.Bookmarks)-1]
			writeXbel() //nolint:errcheck
			return
		}
	}
}

// All returns all the bookmarks in a map of URLs to names.
// It also returns a slice of map keys, sorted so that the map *values*
// are in alphabetical order, with case ignored.
func All() (map[string]string, []string) {
	bkmksMap := make(map[string]string)

	inverted := make(map[string]string)          // Holds inverted map, name->URL
	names := make([]string, len(data.Bookmarks)) // Holds bookmark names, for sorting
	keys := make([]string, len(data.Bookmarks))  // Final sorted keys (URLs), for returning at the end

	for i, bkmk := range data.Bookmarks {
		bkmksMap[bkmk.URL] = bkmk.Name
		inverted[bkmk.Name] = bkmk.URL
		names[i] = bkmk.Name
	}

	// Sort, then turn back into URL keys
	sort.Strings(names)
	for i, name := range names {
		keys[i] = inverted[name]
	}

	return bkmksMap, keys
}
