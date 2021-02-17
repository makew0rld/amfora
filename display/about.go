package display

import (
	"fmt"

	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
)

var aboutPage structs.Page
var versionPage structs.Page
var licensePage structs.Page
var thanksPage structs.Page

func aboutInit(version, commit, builtBy string) {
	aboutPage = createAboutPage("about:about", `# Internal Pages

=> about:bookmarks
=> about:subscriptions
=> about:manage-subscriptions
=> about:newtab
=> about:version
=> about:license
=> about:thanks
`)
	versionPage = createAboutPage("about:version",
		fmt.Sprintf(
			"# Amfora Version Info\n\nAmfora:   %s\nCommit:   %s\nBuilt by: %s",
			version, commit, builtBy,
		),
	)
	licensePage = createAboutPage("about:license", string(license))
	thanksPage = createAboutPage("about:thanks", string(thanks))
}

func createAboutPage(url string, content string) structs.Page {
	renderContent, links := renderer.RenderGemini(content, textWidth(), false)
	return structs.Page{
		Raw:       content,
		Content:   renderContent,
		Links:     links,
		URL:       url,
		Width:     -1, // Force reformatting on first display
		Mediatype: structs.TextGemini,
	}
}
