package display

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/structs"
)

// getSafeDownloadName is used by downloads.go only.
// It returns a modified name that is unique for the downloads folder.
// This way duplicate saved files will not overwrite each other.
//
// lastDot should be set to true if the number added to the name should come before
// the last dot in the filename instead of the first.
//
// n should be set to 0, it is used for recursiveness.
func getSafeDownloadName(name string, lastDot bool, n int) (string, error) {
	// newName("test.txt", 3) -> "test(3).txt"
	newName := func() string {
		if n <= 0 {
			return name
		}
		if lastDot {
			ext := filepath.Ext(name)
			return strings.TrimSuffix(name, ext) + "(" + strconv.Itoa(n) + ")" + ext
		} else {
			idx := strings.Index(name, ".")
			if idx == -1 {
				return name + "(" + strconv.Itoa(n) + ")"
			}
			return name[:idx] + "(" + strconv.Itoa(n) + ")" + name[idx:]
		}
	}

	d, err := os.Open(config.DownloadsDir)
	if err != nil {
		return "", err
	}
	files, err := d.Readdirnames(-1)
	if err != nil {
		d.Close()
		return "", err
	}

	nn := newName()
	for i := range files {
		if nn == files[i] {
			d.Close()
			return getSafeDownloadName(name, lastDot, n+1)
		}
	}
	d.Close()
	return nn, nil // Name doesn't exist already
}

// downloadPage saves the passed Page to a file.
// It returns the saved path and an error.
// It always cleans up, so if an error is returned there is no file saved
func downloadPage(p *structs.Page) (string, error) {
	// Figure out file name
	var name string
	var err error
	parsed, _ := url.Parse(p.Url)
	if parsed.Path == "" || path.Base(parsed.Path) == "/" {
		// No file, just the root domain
		if p.Mediatype == structs.TextGemini {
			name, err = getSafeDownloadName(parsed.Hostname()+".gmi", true, 0)
			if err != nil {
				return "", err
			}
		} else {
			name, err = getSafeDownloadName(parsed.Hostname()+".txt", true, 0)
			if err != nil {
				return "", err
			}
		}
	} else {
		// There's a specific file
		name = path.Base(parsed.Path)
		if p.Mediatype == structs.TextGemini && !strings.HasSuffix(name, ".gmi") && !strings.HasSuffix(name, ".gemini") {
			name += ".gmi"
		}
		name, err = getSafeDownloadName(name, false, 0)
		if err != nil {
			return "", err
		}
	}
	savePath := filepath.Join(config.DownloadsDir, name)
	err = ioutil.WriteFile(savePath, []byte(p.Raw), 0644)
	if err != nil {
		// Just in case
		os.Remove(savePath)
		return "", err
	}
	return savePath, err
}
