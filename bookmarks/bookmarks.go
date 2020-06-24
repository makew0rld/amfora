package bookmarks

import (
	"encoding/base32"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
)

var bkmkStore = config.BkmkStore

// bkmkKey returns the viper key for the given bookmark URL.
// Note that URLs are the keys, NOT the bookmark name.
func bkmkKey(url string) string {
	// Keys are base32 encoded URLs to prevent any bad chars like periods from being used
	return "bookmarks." + base32.StdEncoding.EncodeToString([]byte(url))
}

func Set(url, name string) {
	bkmkStore.Set(bkmkKey(url), name)
	bkmkStore.WriteConfig()
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
	bkmkStore.WriteConfig()
}

// All returns all the bookmarks in a map of URLs to names.
func All() map[string]string {
	ret := make(map[string]string)

	bkmksMap, ok := bkmkStore.AllSettings()["bookmarks"].(map[string]interface{})
	if !ok {
		// No bookmarks stored yet, return empty map
		return ret
	}
	for b32Url, name := range bkmksMap {
		if n, ok := name.(string); n == "" || !ok {
			// name is not a string, or it's empty - ignore
			continue
		}
		url, err := base32.StdEncoding.DecodeString(strings.ToUpper(b32Url))
		if err != nil {
			// This would only happen if a user messed around with the bookmarks file
			continue
		}
		ret[string(url)] = name.(string)
	}
	return ret
}
