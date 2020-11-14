package display

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/spf13/viper"
)

// handleFile handles urls using file:// protocol
func handleFile(u string) (*structs.Page, bool) {
	page := &structs.Page{}

	filename := filepath.FromSlash(strings.TrimPrefix(u, "file://"))

	fi, err := os.Stat(filename)
	if err != nil {
		Error("File Error", "Cannot open local file: "+err.Error())
		return page, false
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		return createDirectoryListing(u)
	case mode.IsRegular():
		if fi.Size() > viper.GetInt64("a-general.page_max_size") {
			Error("File Error", "Cannot open local file, exceeds page max size")
			return page, false
		}

		file, err := os.Open(filename)
		if err != nil {
			Error("File Error", "Cannot open local file: "+err.Error())
			return page, false
		}
		defer file.Close()

		// Read first bytes, to check if plaintext
		buf := make([]byte, 32)
		_, err = file.Read(buf)
		if err != nil && err != io.EOF {
			Error("File Error", "Error reading file: "+err.Error())
			return page, false
		}

		if !utf8.Valid(buf) {
			Error("File Error", "Cannot open local file, looks like a binary.")
			return page, false
		}

		// Looks like plaintext, keep reading
		content := string(buf)
		for {
			_, err := file.Read(buf)
			if err != nil {
				if err != io.EOF {
					Error("File Error", "Cannot open local file: "+err.Error())
					return page, false
				}
				break
			}
			content += string(buf)
		}
		if strings.HasSuffix(u, ".gmi") || strings.HasSuffix(u, ".gemini") {
			rendered, links := renderer.RenderGemini(content, textWidth(), leftMargin(), false)
			page = &structs.Page{
				Mediatype: structs.TextGemini,
				URL:       u,
				Raw:       content,
				Content:   rendered,
				Links:     links,
			}
		} else {
			page = &structs.Page{
				Mediatype: structs.TextPlain,
				URL:       u,
				Raw:       content,
				Content:   renderer.RenderPlainText(content, leftMargin()),
				Links:     []string{},
			}
		}
	}
	return page, true
}

// createDirectoryListing creates a text/gemini page for a directory
// that lists all the files as links.
func createDirectoryListing(u string) (*structs.Page, bool) {
	page := &structs.Page{}
	filename := filepath.FromSlash(strings.TrimPrefix(u, "file://"))
	files, err := ioutil.ReadDir(filename)
	if err != nil {
		Error("Directory error", "Cannot open local directory: "+err.Error())
		return page, false
	}
	content := "Index of " + filename + "\n"
	content += "=> ../ ../\n"
	for _, f := range files {
		separator := ""
		if f.IsDir() {
			separator = "/"
		}
		content += fmt.Sprintf("=> %s%s %s%s\n", f.Name(), separator, f.Name(), separator)
	}

	rendered, links := renderer.RenderGemini(content, textWidth(), leftMargin(), false)
	page = &structs.Page{
		Mediatype: structs.TextGemini,
		URL:       u,
		Raw:       content,
		Content:   rendered,
		Links:     links,
	}
	return page, true
}
