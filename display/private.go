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

// pageUp scrolls up 75% of the height of the terminal, like Bombadillo.
func pageUp() {
	row, col := tabViews[curTab].GetScrollOffset()
	tabViews[curTab].ScrollTo(row-(termH/4)*3, col)
}

// pageDown scrolls down 75% of the height of the terminal, like Bombadillo.
func pageDown() {
	row, col := tabViews[curTab].GetScrollOffset()
	tabViews[curTab].ScrollTo(row+(termH/4)*3, col)
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

// pathEscape is the same as url.PathEscape, but it also replaces the +.
func pathEscape(path string) string {
	return strings.ReplaceAll(url.PathEscape(path), "+", "%2B")
}

// tabHasContent returns true when the current tab has a page being displayed.
// The most likely situation where false would be returned is when the default
// new tab content is being displayed.
func tabHasContent() bool {
	if curTab < 0 {
		return false
	}
	if len(tabViews) < curTab {
		// There isn't a TextView for the current tab number
		return false
	}
	if tabMap[curTab].Url == "" {
		// Likely the default content page
		return false
	}
	if strings.HasPrefix(tabMap[curTab].Url, "about:") {
		return false
	}

	_, ok := tabMap[curTab]
	return ok // If there's a page, return true
}

// saveScroll saves where in the page the user was.
// It should be used whenever moving from one page to another.
func saveScroll() {
	// It will also be saved in the cache because the cache uses the same pointer
	row, col := tabViews[curTab].GetScrollOffset()
	tabMap[curTab].Row = row
	tabMap[curTab].Column = col
}

// applyScroll applies the saved scroll values to the current page and tab.
// It should only be used when going backward and forward, not when
// loading a new page (that might have scroll vals cached anyway).
func applyScroll() {
	tabViews[curTab].ScrollTo(tabMap[curTab].Row, tabMap[curTab].Column)
}

// resolveRelLink returns an absolute link for the given absolute link and relative one.
// It also returns an error if it could not resolve the links, which should be displayed
// to the user.
func resolveRelLink(prev, next string) (string, error) {
	if !tabHasContent() {
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
func followLink(prev, next string) {

	// Copied from URL()
	if next == "about:bookmarks" {
		Bookmarks()
		addToHist("about:bookmarks")
		return
	}
	if strings.HasPrefix(next, "about:") {
		Error("Error", "Not a valid 'about:' URL for linking")
		return
	}

	if tabHasContent() {
		saveScroll() // Likely called later on, it's here just in case
		nextURL, err := resolveRelLink(prev, next)
		if err != nil {
			Error("URL Error", err.Error())
			return
		}
		go func() {
			final, displayed := handleURL(nextURL)
			if displayed {
				addToHist(final)
			}
		}()
		return
	}
	// No content on current tab, so the "prev" URL is not valid.
	// An example is the about:newtab page
	_, err := url.Parse(next)
	if err != nil {
		Error("URL Error", "Link URL could not be parsed")
		return
	}
	go func() {
		final, displayed := handleURL(next)
		if displayed {
			addToHist(final)
		}
	}()
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
	// Links are not recorded because they won't change
	rendered, _ := renderer.RenderGemini(p.Raw, textWidth(), leftMargin())
	p.Content = rendered
	p.Width = termW
}

// reformatAndDisplayPage is for reformatting a page that is already being displayed.
// setPage should be used when a page is being loaded for the first time.
func reformatAndDisplayPage(p *structs.Page) {
	saveScroll()
	reformatPage(tabMap[curTab])
	tabViews[curTab].SetText(tabMap[curTab].Content)
	applyScroll() // Go back to where you were, roughly
}

// setPage displays a Page on the current tab.
func setPage(p *structs.Page) {
	saveScroll() // Save the scroll of the previous page

	// Make sure the page content is fitted to the terminal every time it's displayed
	reformatPage(p)

	// Change page on screen
	tabMap[curTab] = p
	tabViews[curTab].SetText(p.Content)
	tabViews[curTab].Highlight("") // Turn off highlights
	tabViews[curTab].ScrollToBeginning()

	// Setup display
	App.SetFocus(tabViews[curTab])
	bottomBar.SetLabel("")
	bottomBar.SetText(p.Url)
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
func handleURL(u string) (string, bool) {
	defer App.Draw() // Just in case

	App.SetFocus(tabViews[curTab])

	// To allow linking to the bookmarks page, and history browsing
	if u == "about:bookmarks" {
		Bookmarks()
		return "about:bookmarks", true
	}

	u = normalizeURL(u)

	parsed, err := url.Parse(u)
	if err != nil {
		Error("URL Error", err.Error())
		bottomBar.SetText(tabMap[curTab].Url)
		return "", false
	}

	if strings.HasPrefix(u, "http") {
		switch strings.TrimSpace(viper.GetString("a-general.http")) {
		case "", "off":
			Info("Opening HTTP URLs is turned off.")
		case "default":
			s, err := webbrowser.Open(u)
			if err != nil {
				Error("Webbrowser Error", err.Error())
			} else {
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
		bottomBar.SetText(tabMap[curTab].Url)
		return "", false
	}
	if !strings.HasPrefix(u, "gemini") {
		Error("Protocol Error", "Only gemini and HTTP are supported. URL was "+u)
		bottomBar.SetText(tabMap[curTab].Url)
		return "", false
	}
	// Gemini URL

	// Load page from cache if possible
	page, ok := cache.Get(u)
	if ok {
		setPage(page)
		return u, true
	}
	// Otherwise download it
	bottomBar.SetText("Loading...")
	App.Draw()

	res, err := client.Fetch(u)
	if err == client.ErrTofu {
		if Tofu(parsed.Host) {
			// They want to continue anyway
			client.ResetTofuEntry(parsed.Hostname(), parsed.Port(), res.Cert)
			// Response can be used further down, no need to reload
		} else {
			// They don't want to continue
			// Set the bar back to original URL
			bottomBar.SetText(tabMap[curTab].Url)
			return "", false
		}
	} else if err != nil {
		Error("URL Fetch Error", err.Error())
		// Set the bar back to original URL
		bottomBar.SetText(tabMap[curTab].Url)
		return "", false
	}
	if renderer.CanDisplay(res) {
		page, err := renderer.MakePage(u, res, textWidth(), leftMargin())
		page.Width = termW
		if err != nil {
			Error("Page Error", "Issuing creating page: "+err.Error())
			// Set the bar back to original URL
			bottomBar.SetText(tabMap[curTab].Url)
			return "", false
		}
		cache.Add(page)
		setPage(page)
		return u, true
	}
	// Not displayable
	// Could be a non 20 (or 21) status code, or a different kind of document

	// Set the bar back to original URL
	bottomBar.SetText(tabMap[curTab].Url)
	App.Draw()

	// Handle each status code
	switch gemini.SimplifyStatus(res.Status) {
	case 10:
		userInput, ok := Input(res.Meta)
		if ok {
			// Make another request with the query string added
			// + chars are replaced because PathEscape doesn't do that
			parsed.RawQuery = pathEscape(userInput)
			if len(parsed.String()) > 1024 {
				// 1024 is the max size for URLs in the spec
				Error("Input Error", "URL for that input would be too long.")
				return "", false
			}
			return handleURL(parsed.String())
		}
		return "", false
	case 30:
		parsedMeta, err := url.Parse(res.Meta)
		if err != nil {
			Error("Redirect Error", "Invalid URL: "+err.Error())
			return "", false
		}
		redir := parsed.ResolveReference(parsedMeta).String()

		if YesNo("Follow redirect?\n" + redir) {
			return handleURL(redir)
		}
		return "", false
	case 40:
		Error("Temporary Failure", cview.Escape(res.Meta)) // Escaped just in case, to not allow malicious meta strings
		return "", false
	case 50:
		Error("Permanent Failure", cview.Escape(res.Meta))
		return "", false
	case 60:
		Info("The server requested a certificate. Cert handling is coming to Amfora soon!")
		return "", false
	}
	// Status code 20, but not a document that can be displayed
	yes := YesNo("This type of file can't be displayed. Downloading will be implemented soon. Would like to open the file in a HTTPS proxy for now?")
	if yes {
		// Open in mozz's proxy
		portalURL := u
		if parsed.RawQuery != "" {
			// Remove query and add encoded version on the end
			query := parsed.RawQuery
			parsed.RawQuery = ""
			portalURL = parsed.String() + "%3F" + query
		}
		portalURL = strings.TrimPrefix(portalURL, "gemini://") + "?raw=1"

		s, err := webbrowser.Open("https://portal.mozz.us/gemini/" + portalURL)
		if err != nil {
			Error("Webbrowser Error", err.Error())
		} else {
			Info(s)
		}
		App.Draw()
	}
	return "", false
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
