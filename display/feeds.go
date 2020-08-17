package display

import (
	"fmt"
	"strings"

	"github.com/makeworld-the-better-one/amfora/feeds"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
)

var feedPageRaw = "# Feeds & Pages\n\nUpdates" + strings.Repeat(" ", 80-25) + "[Newest -> Oldest]\n" +
	strings.Repeat("-", 80) + "\n\n"

// Feeds displays the feeds page on the current tab.
func Feeds(t *tab) {
	// TODO; Decide about date in local time vs UTC
	// TODO: Cache

	pe := feeds.GetPageEntries()

	curDay := time.Time.Round(time.Day)

	for _, entry := range pe.Entries {
		if entry.Published.Round(time.Day).After(curDay) {
			// This post is on a new day, add a day header
			curDay := entry.Published.Round(time.Day)
			feedPageRaw += fmt.Sprintf("\n## %s\n\n", curDay.Format("Jan 02, 2006"))
		}
		feedPageRaw += fmt.Sprintf("=>%s %s - %s\n", entry.URL, entry.Author, entry.Title)
	}

	content, links := renderer.RenderGemini(feedPageRaw, textWidth(), leftMargin())
	page := structs.Page{
		Raw:       feedPageRaw,
		Content:   content,
		Links:     links,
		Url:       "about:feeds",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	setPage(t, &page)
	t.applyBottomBar()
}
