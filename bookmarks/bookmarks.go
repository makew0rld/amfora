package bookmarks

import (
	"encoding/base32"
	"sort"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
)

var bkmkStore = config.BkmkStore

// bkmkKey returns the viper key for the given bookmark URL.
// Note that URLs are the keys, NOT the bookmark name.
func bkmkKey(url string) string {
	// Keys are base32 encoded URLs to prevent any special chars like periods from being used
	return "bookmarks." + base32.StdEncoding.EncodeToString([]byte(url))
}

func Set(url, name string) {
	bkmkStore.Set(bkmkKey(url), name)
	bkmkStore.WriteConfig() //nolint:errcheck
}

// Get returns the NAME of the bookmark, given the URL.
// It also returns a bool indicating whether it exists.
func Get(url string) (string, bool) {
	name := bkmkStore.GetString(bkmkKey(url))
	return name, name != ""
}

func Remove(url string) {
	// XXX: Viper can't actually delete keys, which means the bookmarks file might get clouded
	// with non-entries over time.
	bkmkStore.Set(bkmkKey(url), "")
	bkmkStore.WriteConfig() //nolint:errcheck
}

// All returns all the bookmarks in a map of URLs to names.
// It also returns a slice of map keys, sorted so that the map *values*
// are in alphabetical order, with case ignored.
func All() (map[string]string, []string) {
	bkmks := make(map[string]string)

	bkmksMap, ok := bkmkStore.AllSettings()["bookmarks"].(map[string]interface{})
	if !ok {
		// No bookmarks stored yet, return empty map
		return bkmks, []string{}
	}

	inverted := make(map[string]string)       // Holds inverted map, name->URL
	names := make([]string, 0, len(bkmksMap)) // Holds bookmark names, for sorting
	keys := make([]string, 0, len(bkmksMap))  // Final sorted keys (URLs), for returning at the end

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
		bkmks[string(url)] = name.(string)
		inverted[name.(string)] = string(url)
		names = append(names, name.(string))
	}
	// Sort, then turn back into URL keys
	sort.Strings(names)
	for _, name := range names {
		keys = append(keys, inverted[name])
	}

	return bkmks, keys
}
