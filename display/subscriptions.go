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
	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/amfora/subscriptions"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

var subscriptionPageUpdated time.Time

// toLocalDay truncates the provided time to a date only,
// but converts to the local time first.
func toLocalDay(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Subscriptions displays the subscriptions page on the current tab.
func Subscriptions(t *tab) {
	logger.Log.Println("display.Subscriptions called")

	// Retrieve cached version if there hasn't been any updates
	p, ok := cache.GetPage("about:subscriptions")
	if subscriptionPageUpdated.After(subscriptions.LastUpdated) && ok {
		logger.Log.Println("using cached subscriptions page")
		setPage(t, p)
		t.applyBottomBar()
		return
	}

	logger.Log.Println("started rendering subscriptions page")

	rawPage := "# Subscriptions\n\n" +
		"See the help (by pressing ?) for details on how to use this page.\n\n" +
		"If you just opened Amfora then updates will appear incrementally. Reload the page to see them.\n\n" +
		"=> about:manage-subscriptions Manage subscriptions\n"

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

	pe := subscriptions.GetPageEntries()

	for _, entry := range pe.Entries { // From new to old
		// Convert to local time, remove sub-day info
		pub := toLocalDay(entry.Published)

		if pub.Before(curDay) {
			// This post is on a new day, add a day header
			curDay = pub
			rawPage += fmt.Sprintf("\n## %s\n\n", curDay.Format("Jan 02, 2006"))
		}
		if entry.Title == "" || entry.Title == "/" {
			// Just put author/title
			// Mainly used for when you're tracking the root domain of a site
			rawPage += fmt.Sprintf("=>%s %s\n", entry.URL, entry.Prefix)
		} else {
			// Include title and dash
			rawPage += fmt.Sprintf("=>%s %s - %s\n", entry.URL, entry.Prefix, entry.Title)
		}
	}

	content, links := renderer.RenderGemini(rawPage, textWidth(), leftMargin(), false)
	page := structs.Page{
		Raw:       rawPage,
		Content:   content,
		Links:     links,
		URL:       "about:subscriptions",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	go cache.AddPage(&page)
	setPage(t, &page)
	t.applyBottomBar()

	subscriptionPageUpdated = time.Now()

	logger.Log.Println("done rendering subscriptions page")
}

// ManageSubscriptions displays the subscription managing page in
// the current tab. `u` is the URL entered by the user.
func ManageSubscriptions(t *tab, u string) {
	if len(u) > 27 && u[:27] == "about:manage-subscriptions?" {
		// There's a query string, aka a URL to unsubscribe from
		manageSubscriptionQuery(t, u)
		return
	}

	rawPage := "# Manage Subscriptions\n\n" +
		"Below is list of URLs, both feeds and pages. Navigate to the link to unsubscribe from that feed or page.\n\n"

	for _, u2 := range subscriptions.AllURLS() {
		rawPage += fmt.Sprintf(
			"=>%s %s\n",
			"about:manage-subscriptions?"+gemini.QueryEscape(u2),
			u2,
		)
	}

	content, links := renderer.RenderGemini(rawPage, textWidth(), leftMargin(), false)
	page := structs.Page{
		Raw:       rawPage,
		Content:   content,
		Links:     links,
		URL:       "about:manage-subscriptions",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	go cache.AddPage(&page)
	setPage(t, &page)
	t.applyBottomBar()
}

func manageSubscriptionQuery(t *tab, u string) {
	sub, err := gemini.QueryUnescape(u[27:])
	if err != nil {
		Error("URL Error", "Invalid query string: "+err.Error())
		return
	}

	err = subscriptions.Remove(sub)
	if err != nil {
		ManageSubscriptions(t, "about:manage-subscriptions") // Reload
		Error("Save Error", "Error saving the unsubscription to disk: "+err.Error())
		return
	}
	ManageSubscriptions(t, "about:manage-subscriptions") // Reload
	Info("Unsubscribed from " + sub)
}

// openSubscriptionModal displays the "Add subscription" modal
// It returns whether the user wanted to subscribe to feed/page.
// The subscribed arg specifies whether this feed/page is already
// subscribed to.
func openSubscriptionModal(validFeed, subscribed bool) bool {
	logger.Log.Println("display.openFeedModal called")
	// Reuses yesNoModal

	if viper.GetBool("a-general.color") {
		yesNoModal.
			SetBackgroundColor(config.GetColor("subscription_modal_bg")).
			SetTextColor(config.GetColor("subscription_modal_text"))
		yesNoModal.GetFrame().
			SetBorderColor(config.GetColor("subscription_modal_text")).
			SetTitleColor(config.GetColor("subscription_modal_text"))
	} else {
		yesNoModal.
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		yesNoModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
	}
	if validFeed {
		yesNoModal.GetFrame().SetTitle("Feed Subscription")
		if subscribed {
			yesNoModal.SetText("You are already subscribed to this feed. Would you like to manually update it?")
		} else {
			yesNoModal.SetText("Would you like to subscribe to this feed?")
		}
	} else {
		yesNoModal.GetFrame().SetTitle("Page Subscription")
		if subscribed {
			yesNoModal.SetText("You are already subscribed to this page. Would you like to manually update it?")
		} else {
			yesNoModal.SetText("Would you like to subscribe to this page?")
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

// getFeedFromPage is like subscriptions.GetFeed but takes a structs.Page as input.
func getFeedFromPage(p *structs.Page) (*gofeed.Feed, bool) {
	parsed, _ := url.Parse(p.URL)
	filename := path.Base(parsed.Path)
	r := strings.NewReader(p.Raw)
	return subscriptions.GetFeed(p.RawMediatype, filename, r)
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

	if openSubscriptionModal(true, tracked) {
		err := subscriptions.AddFeed(u, feed)
		if err != nil {
			Error("Feed Error", err.Error())
		}
		return true
	}
	return false
}

// addFeed goes through the process of subscribing to the current page/feed.
// It is the high-level way of doing it. It should be called in a goroutine.
func addSubscription() {
	logger.Log.Println("display.addSubscription called")

	t := tabs[curTab]
	p := t.page

	if !t.hasContent() {
		// It's an about: page, or a malformed one
		return
	}

	feed, isFeed := getFeedFromPage(p)
	tracked := subscriptions.IsSubscribed(p.URL)

	if openSubscriptionModal(isFeed, tracked) {
		var err error

		if isFeed {
			err = subscriptions.AddFeed(p.URL, feed)
		} else {
			err = subscriptions.AddPage(p.URL, strings.NewReader(p.Raw))
		}

		if err != nil {
			Error("Feed/Page Error", err.Error())
		}
	}
}
