package display

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/amfora/subscriptions"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

// Map page number (zero-indexed) to the time it was made at.
// This allows for caching the pages until there's an update.
var subscriptionPageUpdated = make(map[int]time.Time)

// toLocalDay truncates the provided time to a date only,
// but converts to the local time first.
func toLocalDay(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Subscriptions displays the subscriptions page on the current tab.
func Subscriptions(t *tab, u string) string {
	pageN := 0 // Pages are zero-indexed internally

	// Correct URL if query string exists
	// The only valid query string is an int above 1.
	// Anything "redirects" to the first page, with no query string.
	// This is done over just serving the first page content for
	// invalid query strings so that there won't be duplicate caches.
	correctURL := func(u2 string) string {
		if len(u2) > 20 && u2[:20] == "about:subscriptions?" {
			query, err := gemini.QueryUnescape(u2[20:])
			if err != nil {
				return "about:subscriptions"
			}
			// Valid query string
			i, err := strconv.Atoi(query)
			if err != nil {
				// Not an int
				return "about:subscriptions"
			}
			if i < 2 {
				return "about:subscriptions"
			}
			// Valid int above 1
			pageN = i - 1 // Pages are zero-indexed internally
			return u2
		}
		return u2
	}
	u = correctURL(u)

	// Retrieve cached version if there hasn't been any updates
	p, ok := cache.GetPage(u)
	if subscriptionPageUpdated[pageN].After(subscriptions.LastUpdated) && ok {
		setPage(t, p)
		t.applyBottomBar()
		return u
	}

	pe := subscriptions.GetPageEntries()

	// Figure out where the entries for this page start, if at all.
	epp := viper.GetInt("subscriptions.entries_per_page")
	if epp <= 0 {
		epp = 1
	}
	start := pageN * epp // Index of the first page entry to be displayed
	end := start + epp
	if end > len(pe.Entries) {
		end = len(pe.Entries)
	}

	var rawPage string
	if pageN == 0 {
		rawPage = "# Subscriptions\n\n" + rawPage
	} else {
		rawPage = fmt.Sprintf("# Subscriptions (page %d)\n\n", pageN+1) + rawPage
	}

	if start > len(pe.Entries)-1 && len(pe.Entries) != 0 {
		// The page is out of range, doesn't exist
		rawPage += "This page does not exist.\n\n=> about:subscriptions Subscriptions\n"
	} else {
		// Render page

		rawPage += "You can use Ctrl-X to subscribe to a page, or to an Atom/RSS/JSON feed. See the online wiki for more.\n" +
			"If you just opened Amfora then updates may appear incrementally. Reload the page to see them.\n\n" +
			"=> about:manage-subscriptions Manage subscriptions\n\n"

		// curDay represents what day of posts the loop is on.
		// It only goes backwards in time.
		// Its initial setting means:
		// Only display posts older than 26 hours in the future, nothing further in the future.
		//
		// 26 hours was chosen because it is the largest timezone difference
		// currently in the world. Posts may be dated in the future
		// due to software bugs, where the local user's date is used, but
		// the UTC timezone is specified. Gemfeed does this at the time of
		// writing, but will not after #3 gets merged on its repo. Still,
		// the older version will be used for a while.
		curDay := toLocalDay(time.Now()).Add(26 * time.Hour)

		for _, entry := range pe.Entries[start:end] { // From new to old
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

		if pageN == 0 && len(pe.Entries) > epp {
			// First page, and there's more than can fit
			rawPage += "\n\n=> about:subscriptions?2 Next Page\n"
		} else if pageN > 0 {
			// A later page
			rawPage += fmt.Sprintf(
				"\n\n=> about:subscriptions?%d Previous Page\n",
				pageN, // pageN is zero-indexed but the query string is one-indexed
			)
			if end != len(pe.Entries) {
				// There's more
				rawPage += fmt.Sprintf("=> about:subscriptions?%d Next Page\n", pageN+2)
			}
		}
	}

	content, links := renderer.RenderGemini(rawPage, textWidth(), false)
	page := structs.Page{
		Raw:       rawPage,
		Content:   content,
		Links:     links,
		URL:       u,
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	go cache.AddPage(&page)
	setPage(t, &page)
	t.applyBottomBar()

	subscriptionPageUpdated[pageN] = time.Now()

	return u
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
		"Below is list of URLs you are subscribed to, both feeds and pages. " +
		"Navigate to the link to unsubscribe from that feed or page.\n\n"

	urls := subscriptions.AllURLS()
	sort.Strings(urls)

	for _, u2 := range urls {
		rawPage += fmt.Sprintf(
			"=>%s %s\n",
			"about:manage-subscriptions?"+gemini.QueryEscape(u2),
			u2,
		)
	}

	content, links := renderer.RenderGemini(rawPage, textWidth(), false)
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
	// Reuses yesNoModal

	if viper.GetBool("a-general.color") {
		m := yesNoModal
		m.SetBackgroundColor(config.GetColor("subscription_modal_bg"))
		m.SetTextColor(config.GetColor("subscription_modal_text"))
		frame := yesNoModal.GetFrame()
		frame.SetBorderColor(config.GetColor("subscription_modal_text"))
		frame.SetTitleColor(config.GetColor("subscription_modal_text"))
	} else {
		m := yesNoModal
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		frame := yesNoModal.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)
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

	panels.ShowPanel("yesno")
	panels.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	panels.HidePanel("yesno")
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
