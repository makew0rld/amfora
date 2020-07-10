package display

import (
	"errors"
	"net/url"
	"os/exec"
	"strings"

	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/amfora/webbrowser"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// This file contains the functions that aren't part of the public API.

// isValidTab indicates whether the passed tab is still being used, even if it's not currently displayed.
func isValidTab(t *tab) bool {
	tempTabs := tabs
	for i := range tempTabs {
		if tempTabs[i] == t {
			return true
		}
	}
	return false
}

func leftMargin() int {
	return int(float64(termW) * viper.GetFloat64("a-general.left_margin"))
}

func textWidth() int {
	if termW <= 0 {
		// This prevent a flash of 1-column text on startup, when the terminal
		// width hasn't been initialized.
		return viper.GetInt("a-general.max_width")
	}

	rightMargin := leftMargin()
	if leftMargin() > 10 {
		// 10 is the max right margin
		rightMargin = 10
	}

	max := termW - leftMargin() - rightMargin
	if max < viper.GetInt("a-general.max_width") {
		return max
	}
	return viper.GetInt("a-general.max_width")
}

// queryEscape is the same as url.PathEscape, but it also replaces the +.
// This is because Gemini requires percent-escaping for queries.
func queryEscape(path string) string {
	return strings.ReplaceAll(url.PathEscape(path), "+", "%2B")
}

// resolveRelLink returns an absolute link for the given absolute link and relative one.
// It also returns an error if it could not resolve the links, which should be displayed
// to the user.
func resolveRelLink(t *tab, prev, next string) (string, error) {
	if !t.hasContent() {
		return next, nil
	}

	prevParsed, _ := url.Parse(prev)
	nextParsed, err := url.Parse(next)
	if err != nil {
		return "", errors.New("link URL could not be parsed")
	}
	return prevParsed.ResolveReference(nextParsed).String(), nil
}

// followLink should be used when the user "clicks" a link on a page.
// Not when a URL is opened on a new tab for the first time.
// It will handle setting the bottomBar.
func followLink(t *tab, prev, next string) {

	// Copied from URL()
	if next == "about:bookmarks" {
		Bookmarks(t)
		t.addToHistory("about:bookmarks")
		return
	}
	if strings.HasPrefix(next, "about:") {
		Error("Error", "Not a valid 'about:' URL for linking")
		return
	}

	if t.hasContent() {
		t.saveScroll() // Likely called later on, it's here just in case
		nextURL, err := resolveRelLink(t, prev, next)
		if err != nil {
			Error("URL Error", err.Error())
			return
		}
		go goURL(t, nextURL)
		return
	}
	// No content on current tab, so the "prev" URL is not valid.
	// An example is the about:newtab page
	_, err := url.Parse(next)
	if err != nil {
		Error("URL Error", "Link URL could not be parsed")
		return
	}
	go goURL(t, next)
}

// reformatPage will take the raw page content and reformat it according to the current terminal dimensions.
// It should be called when the terminal size changes.
// It will not waste resources if the passed page is already fitted to the current terminal width, and can be
// called safely even when the page might be already formatted properly.
func reformatPage(p *structs.Page) {
	if p.Width == termW {
		// No changes to make
		return
	}

	var rendered string
	if p.Mediatype == structs.TextGemini {
		// Links are not recorded because they won't change
		rendered, _ = renderer.RenderGemini(p.Raw, textWidth(), leftMargin())
	} else if p.Mediatype == structs.TextPlain {
		rendered = renderer.RenderPlainText(p.Raw, leftMargin())
	} else if p.Mediatype == structs.TextAnsi {
		rendered = renderer.RenderANSI(p.Raw, leftMargin())
	} else {
		// Rendering this type is not implemented
		return
	}
	p.Content = rendered
	p.Width = termW
}

// reformatPageAndSetView is for reformatting a page that is already being displayed.
// setPage should be used when a page is being loaded for the first time.
func reformatPageAndSetView(t *tab, p *structs.Page) {
	t.saveScroll()
	reformatPage(p)
	t.view.SetText(p.Content)
	t.applyScroll() // Go back to where you were, roughly
}

// setPage displays a Page on the passed tab number.
// The bottomBar is not actually changed in this func
func setPage(t *tab, p *structs.Page) {
	if !isValidTab(t) {
		// Don't waste time reformatting an invalid tab
		return
	}

	t.saveScroll() // Save the scroll of the previous page

	// Make sure the page content is fitted to the terminal every time it's displayed
	reformatPage(p)

	// Change page on screen
	t.page = p
	t.view.SetText(p.Content)
	t.view.Highlight("") // Turn off highlights, other funcs may restore if necessary
	t.view.ScrollToBeginning()

	// Setup display
	App.SetFocus(t.view)

	// Save bottom bar for the tab - other funcs will apply/display it
	t.barLabel = ""
	t.barText = p.Url
}

// handleHTTP is used by handleURL.
// It opens HTTP links and displays Info and Error modals.
func handleHTTP(u string, showInfo bool) {
	switch strings.TrimSpace(viper.GetString("a-general.http")) {
	case "", "off":
		Error("HTTP Error", "Opening HTTP URLs is turned off.")
	case "default":
		s, err := webbrowser.Open(u)
		if err != nil {
			Error("Webbrowser Error", err.Error())
		} else if showInfo {
			Info(s)
		}
	default:
		// The config has a custom command to execute for HTTP URLs
		fields := strings.Fields(viper.GetString("a-general.http"))
		err := exec.Command(fields[0], append(fields[1:], u)...).Start()
		if err != nil {
			Error("HTTP Error", "Error executing custom browser command: "+err.Error())
		}
	}
	App.Draw()
}

// goURL is like handleURL, but takes care of history and the bottomBar.
// It should be preferred over handleURL in most cases.
// It has no return values to be processed.
//
// It should be called in a goroutine.
func goURL(t *tab, u string) {
	final, displayed := handleURL(t, u)
	if displayed {
		t.addToHistory(final)
	}
	if t == tabs[curTab] {
		// Display the bottomBar state that handleURL set
		t.applyBottomBar()
	}
}

// handleURL displays whatever action is needed for the provided URL,
// and applies it to the current tab.
// It loads documents, handles errors, brings up a download prompt, etc.
//
// The string returned is the final URL, if redirects were involved.
// In most cases it will be the same as the passed URL.
// If there is some error, it will return "".
// The second returned item is a bool indicating if page content was displayed.
// It returns false for Errors, other protocols, etc.
//
// The bottomBar is not actually changed in this func, except during loading.
// The func that calls this one should apply the bottomBar values if necessary.
func handleURL(t *tab, u string) (string, bool) {
	defer App.Draw() // Just in case

	// Save for resetting on error
	oldLable := t.barLabel
	oldText := t.barText

	// Custom return function
	ret := func(s string, b bool) (string, bool) {
		if !b {
			// Reset bottomBar if page wasn't loaded
			t.barLabel = oldLable
			t.barText = oldText
		}
		t.mode = tabModeDone
		return s, b
	}

	t.barLabel = ""
	bottomBar.SetLabel("")

	App.SetFocus(t.view)

	// To allow linking to the bookmarks page, and history browsing
	if u == "about:bookmarks" {
		Bookmarks(t)
		return ret("about:bookmarks", true)
	}

	u = normalizeURL(u)

	parsed, err := url.Parse(u)
	if err != nil {
		Error("URL Error", err.Error())
		return ret("", false)
	}

	if strings.HasPrefix(u, "http") {
		handleHTTP(u, true)
		return ret("", false)
	}
	if !strings.HasPrefix(u, "gemini") {
		Error("Protocol Error", "Only gemini and HTTP are supported. URL was "+u)
		return ret("", false)
	}
	// Gemini URL

	// Load page from cache if possible
	page, ok := cache.Get(u)
	if ok {
		setPage(t, page)
		return ret(u, true)
	}
	// Otherwise download it
	bottomBar.SetText("Loading...")
	t.barText = "Loading..." // Save it too, in case the tab switches during loading
	t.mode = tabModeLoading
	App.Draw()

	res, err := client.Fetch(u)

	// Loading may have taken a while, make sure tab is still valid
	if !isValidTab(t) {
		return ret("", false)
	}

	if err == client.ErrTofu {
		if Tofu(parsed.Host) {
			// They want to continue anyway
			client.ResetTofuEntry(parsed.Hostname(), parsed.Port(), res.Cert)
			// Response can be used further down, no need to reload
		} else {
			// They don't want to continue
			return ret("", false)
		}
	} else if err != nil {
		Error("URL Fetch Error", err.Error())
		return ret("", false)
	}
	if renderer.CanDisplay(res) {
		page, err := renderer.MakePage(u, res, textWidth(), leftMargin())
		// Rendering may have taken a while, make sure tab is still valid
		if !isValidTab(t) {
			return ret("", false)
		}

		// Make new request for downloading purposes
		res, clientErr := client.Fetch(u)
		if clientErr != nil && clientErr != client.ErrTofu {
			Error("URL Fetch Error", err.Error())
			return ret("", false)
		}

		if err == renderer.ErrTooLarge {
			go dlChoice("That page is too large. What would you like to do?", u, res)
			return ret("", false)
		}
		if err == renderer.ErrTimedOut {
			go dlChoice("Loading that page timed out. What would you like to do?", u, res)
			return ret("", false)
		}
		if err != nil {
			Error("Page Error", "Issuing creating page: "+err.Error())
			return ret("", false)
		}

		page.Width = termW
		go cache.Add(page)
		setPage(t, page)
		return ret(u, true)
	}
	// Not displayable
	// Could be a non 20 (or 21) status code, or a different kind of document

	// Handle each status code
	switch gemini.SimplifyStatus(res.Status) {
	case 10:
		userInput, ok := Input(res.Meta)
		if ok {
			// Make another request with the query string added
			// + chars are replaced because PathEscape doesn't do that
			parsed.RawQuery = queryEscape(userInput)
			if len(parsed.String()) > 1024 {
				// 1024 is the max size for URLs in the spec
				Error("Input Error", "URL for that input would be too long.")
				return ret("", false)
			}
			return ret(handleURL(t, parsed.String()))
		}
		return ret("", false)
	case 30:
		parsedMeta, err := url.Parse(res.Meta)
		if err != nil {
			Error("Redirect Error", "Invalid URL: "+err.Error())
			return ret("", false)
		}
		redir := parsed.ResolveReference(parsedMeta).String()
		if YesNo("Follow redirect?\n" + redir) {
			return handleURL(t, redir)
		}
		return ret("", false)
	case 40:
		Error("Temporary Failure", cview.Escape(res.Meta))
		return ret("", false)
	case 50:
		Error("Permanent Failure", cview.Escape(res.Meta))
		return ret("", false)
	case 60:
		Info("The server requested a certificate. Cert handling is coming to Amfora soon!")
		return ret("", false)
	}
	// Status code 20, but not a document that can be displayed
	go dlChoice("That file could not be displayed. What would you like to do?", u, res)
	return ret("", false)
}

// normalizeURL attempts to make URLs that are different strings
// but point to the same place all look the same.
//
// Example: gemini://gus.guru:1965/ and //gus.guru/.
// This function will take both output the same URL each time.
//
// The string passed must already be confirmed to be a URL.
// Detection of a search string vs. a URL must happen elsewhere.
//
// It only works with absolute URLs.
func normalizeURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	if !strings.Contains(u, "://") && !strings.HasPrefix(u, "//") {
		// No scheme at all in the URL
		parsed, err = url.Parse("gemini://" + u)
		if err != nil {
			return u
		}
	}
	if parsed.Scheme == "" {
		// Always add scheme
		parsed.Scheme = "gemini"
	} else if parsed.Scheme != "gemini" {
		// Not a gemini URL, nothing to do
		return u
	}

	parsed.User = nil    // No passwords in Gemini
	parsed.Fragment = "" // No fragments either
	if parsed.Port() == "1965" {
		// Always remove default port
		parsed.Host = parsed.Hostname()
	}

	// Add slash to the end of a URL with just a domain
	// gemini://example.com -> gemini://example.com/
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String()
}
