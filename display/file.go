package display

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

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

		mimetype := mime.TypeByExtension(filepath.Ext(filename))
		if strings.HasSuffix(u, ".gmi") || strings.HasSuffix(u, ".gemini") {
			mimetype = "text/gemini"
		}

		if !strings.HasPrefix(mimetype, "text/") {
			Error("File Error", "Cannot open file, unknown mimetype.")
			return page, false
		}

		content, err := ioutil.ReadFile(filename)

		if err != nil {
			Error("File Error", "Cannot open local file: "+err.Error())
			return page, false
		}

		if mimetype == "text/gemini" {
			rendered, links := renderer.RenderGemini(string(content), textWidth(), leftMargin(), false)
			page = &structs.Page{
				Mediatype: structs.TextGemini,
				URL:       u,
				Raw:       string(content),
				Content:   rendered,
				Links:     links,
				Width:     termW,
			}
		} else {
			page = &structs.Page{
				Mediatype: structs.TextPlain,
				URL:       u,
				Raw:       string(content),
				Content:   renderer.RenderPlainText(string(content), leftMargin()),
				Links:     []string{},
				Width:     termW,
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
		Width:     termW,
	}
	return page, true
}
