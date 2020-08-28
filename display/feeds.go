package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/feeds"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
)

var feedPageRaw = "# Feeds & Pages\n\nUpdates" + strings.Repeat(" ", 80-25) + "[Newest -> Oldest]\n" +
	strings.Repeat("-", 80) + "\n\n"

var timeDay = 24 * time.Hour

var feedPageUpdated time.Time

// Feeds displays the feeds page on the current tab.
func Feeds(t *tab) {
	// Retrieve cached version if there hasn't been updates
	p, ok := cache.GetPage("about:feeds")
	if feedPageUpdated == feeds.LastUpdated && ok {
		setPage(t, p)
		t.applyBottomBar()
		return
	}

	pe := feeds.GetPageEntries()

	// curDay represents what day of posts the loop is on.
	// It only goes backwards in time.
	// It's initial setting means:
	// only display posts older than a day in the future.
	curDay := time.Now().Round(timeDay).Add(timeDay)

	for _, entry := range pe.Entries { // From new to old
		// Convert to local time, remove sub-day info
		pub := entry.Published.In(time.Local).Round(timeDay)

		if pub.Before(curDay) {
			// This post is on a new day, add a day header
			curDay := pub
			feedPageRaw += fmt.Sprintf("\n## %s\n\n", curDay.Format("Jan 02, 2006"))
		}
		feedPageRaw += fmt.Sprintf("=>%s %s - %s\n", entry.URL, entry.Author, entry.Title)
	}

	content, links := renderer.RenderGemini(feedPageRaw, textWidth(), leftMargin())
	page := structs.Page{
		Raw:       feedPageRaw,
		Content:   content,
		Links:     links,
		URL:       "about:feeds",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	cache.AddPage(&page)
	setPage(t, &page)
	t.applyBottomBar()

	feedPageUpdated = time.Now()
}

func feedInit() {
	// TODO
}
