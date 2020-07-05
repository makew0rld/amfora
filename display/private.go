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
func resolveRelLink(tab int, prev, next string) (string, error) {
	if !tabs[tab].hasContent() {
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
func followLink(tab int, prev, next string) {

	// Copied from URL()
	if next == "about:bookmarks" {
		Bookmarks()
		tabs[tab].addToHistory("about:bookmarks")
		return
	}
	if strings.HasPrefix(next, "about:") {
		Error("Error", "Not a valid 'about:' URL for linking")
		return
	}

	if tabs[tab].hasContent() {
		tabs[tab].saveScroll() // Likely called later on, it's here just in case
		nextURL, err := resolveRelLink(tab, prev, next)
		if err != nil {
			Error("URL Error", err.Error())
			return
		}
		go goURL(tab, nextURL)
		return
	}
	// No content on current tab, so the "prev" URL is not valid.
	// An example is the about:newtab page
	_, err := url.Parse(next)
	if err != nil {
		Error("URL Error", "Link URL could not be parsed")
		return
	}
	go goURL(tab, next)
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
	} else {
		// Rendering this type is not implemented
		return
	}
	p.Content = rendered
	p.Width = termW
}

// reformatPageAndSetView is for reformatting a page that is already being displayed.
// setPage should be used when a page is being loaded for the first time.
func reformatPageAndSetView(tab int, p *structs.Page) {
	tabs[tab].saveScroll()
	reformatPage(p)
	tabs[tab].view.SetText(p.Content)
	tabs[tab].applyScroll() // Go back to where you were, roughly
}

// setPage displays a Page on the passed tab number.
// The bottomBar is not actually changed in this func
func setPage(tab int, p *structs.Page) {
	tabs[tab].saveScroll() // Save the scroll of the previous page

	// Make sure the page content is fitted to the terminal every time it's displayed
	reformatPage(p)

	// Change page on screen
	tabs[tab].page = p
	tabs[tab].view.SetText(p.Content)
	tabs[tab].view.Highlight("") // Turn off highlights
	tabs[tab].view.ScrollToBeginning()

	// Setup display
	App.SetFocus(tabs[tab].view)

	// Save bottom bar for the tab - TODO: other funcs will apply/display it
	tabs[tab].barLabel = ""
	tabs[tab].barText = p.Url
}

// goURL is like handleURL, but takes care of history and the bottomBar.
// It should be preferred over handleURL in most cases.
// It has no return values to be processed.
//
// It should be called in a goroutine.
func goURL(tab int, u string) {
	final, displayed := handleURL(tab, u)
	if displayed {
		tabs[tab].addToHistory(final)
	}
	if tab == curTab {
		// Display the bottomBar state that handleURL set
		tabs[tab].applyBottomBar()
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
func handleURL(tab int, u string) (string, bool) {
	defer App.Draw() // Just in case

	App.SetFocus(tabs[tab].view)

	// To allow linking to the bookmarks page, and history browsing
	if u == "about:bookmarks" {
		Bookmarks()
		return "about:bookmarks", true
	}

	u = normalizeURL(u)

	parsed, err := url.Parse(u)
	if err != nil {
		Error("URL Error", err.Error())
		tabs[tab].barText = tabs[tab].page.Url
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
		tabs[tab].barText = tabs[tab].page.Url
		return "", false
	}
	if !strings.HasPrefix(u, "gemini") {
		Error("Protocol Error", "Only gemini and HTTP are supported. URL was "+u)
		tabs[tab].barText = tabs[tab].page.Url
		return "", false
	}
	// Gemini URL

	// Load page from cache if possible
	page, ok := cache.Get(u)
	if ok {
		setPage(tab, page)
		return u, true
	}
	// Otherwise download it
	bottomBar.SetText("Loading...")
	tabs[tab].barText = "Loading..." // Save it too, in case the tab switches during loading
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
			tabs[tab].barText = tabs[tab].page.Url
			return "", false
		}
	} else if err != nil {
		Error("URL Fetch Error", err.Error())
		// Set the bar back to original URL
		tabs[tab].barText = tabs[tab].page.Url
		return "", false
	}
	if renderer.CanDisplay(res) {
		page, err := renderer.MakePage(u, res, textWidth(), leftMargin())
		page.Width = termW
		if err != nil {
			Error("Page Error", "Issuing creating page: "+err.Error())
			// Set the bar back to original URL
			tabs[tab].barText = tabs[tab].page.Url
			return "", false
		}
		cache.Add(page)
		setPage(tab, page)
		return u, true
	}
	// Not displayable
	// Could be a non 20 (or 21) status code, or a different kind of document

	// Set the bar back to original URL
	bottomBar.SetText(tabs[curTab].page.Url)
	tabs[tab].barText = tabs[curTab].page.Url
	App.Draw()

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
				return "", false
			}
			return handleURL(tab, parsed.String())
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
			return handleURL(tab, redir)
		}
		return "", false
	case 40:
		Error("Temporary Failure", cview.Escape(res.Meta))
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
