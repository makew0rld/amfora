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

var feedPageUpdated time.Time

// toLocalDay truncates the provided time to a date only,
// but converts to the local time first.
func toLocalDay(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Feeds displays the feeds page on the current tab.
func Feeds(t *tab) {
	// Retrieve cached version if there hasn't been updates
	p, ok := cache.GetPage("about:feeds")
	if feedPageUpdated.After(feeds.LastUpdated) && ok {
		setPage(t, p)
		t.applyBottomBar()
		return
	}

	// curDay represents what day of posts the loop is on.
	// It only goes backwards in time.
	// It's initial setting means:
	// Only display posts older than 6 hours in the future,
	// nothing further in the future.
	curDay := toLocalDay(time.Now()).Add(6 * time.Hour)

	pe := feeds.GetPageEntries()

	for _, entry := range pe.Entries { // From new to old
		// Convert to local time, remove sub-day info
		pub := toLocalDay(entry.Published)

		if pub.Before(curDay) {
			// This post is on a new day, add a day header
			curDay := pub
			feedPageRaw += fmt.Sprintf("\n## %s\n\n", curDay.Format("Jan 02, 2006"))
		}
		feedPageRaw += fmt.Sprintf("=>%s %s - %s\n", entry.URL, entry.Author, entry.Title)
	}

	content, links := renderer.RenderGemini(feedPageRaw, textWidth(), leftMargin(), false)
	page := structs.Page{
		Raw:       feedPageRaw,
		Content:   content,
		Links:     links,
		URL:       "about:feeds",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	go cache.AddPage(&page)
	setPage(t, &page)
	t.applyBottomBar()

	feedPageUpdated = time.Now()
}

func feedInit() {
	// TODO
}
