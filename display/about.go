package display

import (
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
)

var aboutPage structs.Page
var versionPage structs.Page
var licensePage structs.Page
var thanksPage structs.Page

func aboutInit() {
	aboutPage = createAboutPage("about:about", `# Builtin Pages

=> about:bookmarks Your bookmarks
=> about:subscriptions Your subscriptions
=> about:manage-subscriptions Manage your subscriptions
=> about:newtab A new tab
=> about:version Version and build information
=> about:license License and copyright information
=> about:thanks Credits
=> about:about This page
`)
	versionPage = createAboutPage("about:version", "# Amfora Version Info\n\nAmfora:   %s\nCommit:   %s\nBuilt by: %s")
	licensePage = createAboutPage("about:license", string(license))
	thanksPage = createAboutPage("about:thanks", string(thanks))
}

func createAboutPage(url string, content string) structs.Page {
	renderContent, links := renderer.RenderGemini(content, textWidth(), leftMargin(), false)
	return structs.Page{
		Raw:       content,
		Content:   renderContent,
		Links:     links,
		URL:       url,
		Width:     -1, // Force reformatting on first display
		Mediatype: structs.TextGemini,
	}
}
