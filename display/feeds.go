package display

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/feeds"
	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

var feedPageUpdated time.Time

// toLocalDay truncates the provided time to a date only,
// but converts to the local time first.
func toLocalDay(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Feeds displays the feeds page on the current tab.
func Feeds(t *tab) {
	logger.Log.Println("display.Feeds called")

	// Retrieve cached version if there hasn't been any updates
	p, ok := cache.GetPage("about:feeds")
	if feedPageUpdated.After(feeds.LastUpdated) && ok {
		logger.Log.Println("using cached feeds page")
		setPage(t, p)
		t.applyBottomBar()
		return
	}

	logger.Log.Println("started rendering feeds page")

	feedPageRaw := "# Feeds & Pages\n\n" +
		"See the help (by pressing ?) for details on how to use this page.\n\n" +
		"If you just opened Amfora then updates will appear incrementally. Reload the page to see them.\n"

	// curDay represents what day of posts the loop is on.
	// It only goes backwards in time.
	// Its initial setting means:
	// Only display posts older than 26 hours in the future, nothing further in the future.
	//
	// 26 hours was chosen because it is the largest timezone difference
	// currently in the world. Posts may be dated in the future
	// due to software bugs, where the local user's date is used, but
	// the UTC timezone is specified. I believe gemfeed does this.
	curDay := toLocalDay(time.Now()).Add(26 * time.Hour)

	pe := feeds.GetPageEntries()

	for _, entry := range pe.Entries { // From new to old
		// Convert to local time, remove sub-day info
		pub := toLocalDay(entry.Published)

		if pub.Before(curDay) {
			// This post is on a new day, add a day header
			curDay = pub
			feedPageRaw += fmt.Sprintf("\n## %s\n\n", curDay.Format("Jan 02, 2006"))
		}
		if entry.Title == "" || entry.Title == "/" {
			// Just put author/title
			// Mainly used for when you're tracking the root domain of a site
			feedPageRaw += fmt.Sprintf("=>%s %s\n", entry.URL, entry.Prefix)
		} else {
			// Include title and dash
			feedPageRaw += fmt.Sprintf("=>%s %s - %s\n", entry.URL, entry.Prefix, entry.Title)
		}
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

	logger.Log.Println("done rendering feeds page")
}

// openFeedModal displays the "Add feed/page" modal
// It returns whether the user wanted to add the feed/page.
// The tracked arg specifies whether this feed/page is already
// being tracked.
func openFeedModal(validFeed, tracked bool) bool {
	logger.Log.Println("display.openFeedModal called")
	// Reuses yesNoModal

	if viper.GetBool("a-general.color") {
		yesNoModal.
			SetBackgroundColor(config.GetColor("feed_modal_bg")).
			SetTextColor(config.GetColor("feed_modal_text"))
		yesNoModal.GetFrame().
			SetBorderColor(config.GetColor("feed_modal_text")).
			SetTitleColor(config.GetColor("feed_modal_text"))
	} else {
		yesNoModal.
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		yesNoModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
	}
	if validFeed {
		yesNoModal.GetFrame().SetTitle("Feed Tracking")
		if tracked {
			yesNoModal.SetText("This is already being tracked. Would you like to manually update it?")
		} else {
			yesNoModal.SetText("Would you like to start tracking this feed?")
		}
	} else {
		yesNoModal.GetFrame().SetTitle("Page Tracking")
		if tracked {
			yesNoModal.SetText("This is already being tracked. Would you like to manually update it?")
		} else {
			yesNoModal.SetText("Would you like to start tracking this page?")
		}
	}

	tabPages.ShowPage("yesno")
	tabPages.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	App.SetFocus(tabs[curTab].view)
	App.Draw()
	return resp
}

// getFeedFromPage is like feeds.GetFeed but takes a structs.Page as input.
func getFeedFromPage(p *structs.Page) (*gofeed.Feed, bool) {
	parsed, _ := url.Parse(p.URL)
	filename := path.Base(parsed.Path)
	r := strings.NewReader(p.Raw)
	return feeds.GetFeed(p.RawMediatype, filename, r)
}

// addFeedDirect is only for adding feeds, not pages.
// It's for when you already have a feed and know if it's tracked.
// Used mainly by handleURL because it already did a lot of the work.
// It returns a bool indicating whether the user actually wanted to
// add the feed or not.
//
// Like addFeed, it should be called in a goroutine.
func addFeedDirect(u string, feed *gofeed.Feed, tracked bool) bool {
	logger.Log.Println("display.addFeedDirect called")

	if openFeedModal(true, tracked) {
		err := feeds.AddFeed(u, feed)
		if err != nil {
			Error("Feed Error", err.Error())
		}
		return true
	}
	return false
}

// addFeed goes through the process of tracking the current page/feed.
// It is the high-level way of doing it. It should be called in a goroutine.
func addFeed() {
	logger.Log.Println("display.addFeed called")

	t := tabs[curTab]
	p := t.page

	if !t.hasContent() {
		// It's an about: page, or a malformed one
		return
	}

	feed, isFeed := getFeedFromPage(p)
	tracked := feeds.IsTracked(p.URL)

	if openFeedModal(isFeed, tracked) {
		var err error

		if isFeed {
			err = feeds.AddFeed(p.URL, feed)
		} else {
			err = feeds.AddPage(p.URL, strings.NewReader(p.Raw))
		}

		if err != nil {
			Error("Feed/Page Error", err.Error())
		}
	}
}
