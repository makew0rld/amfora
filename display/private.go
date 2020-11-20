package display

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/feeds"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/amfora/webbrowser"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/makeworld-the-better-one/go-isemoji"
	"github.com/spf13/viper"
)

// This file contains the functions that aren't part of the public API.
// The funcs are for network and displaying.

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
	if next == "about:feeds" {
		Feeds(t)
		t.addToHistory("about:feeds")
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

	// TODO: Setup a renderer.RenderFromMediatype func so this isn't needed

	var rendered string
	switch p.Mediatype {
	case structs.TextGemini:
		// Links are not recorded because they won't change
		proxied := true
		if strings.HasPrefix(p.URL, "gemini") || strings.HasPrefix(p.URL, "about") {
			proxied = false
		}
		rendered, _ = renderer.RenderGemini(p.Raw, textWidth(), leftMargin(), proxied)
	case structs.TextPlain:
		rendered = renderer.RenderPlainText(p.Raw, leftMargin())
	case structs.TextAnsi:
		rendered = renderer.RenderANSI(p.Raw, leftMargin())
	default:
		// Rendering this type is not implemented
		return
	}
	p.Content = rendered
	p.Width = termW
}

// reformatPageAndSetView is for reformatting a page that is already being displayed.
// setPage should be used when a page is being loaded for the first time.
func reformatPageAndSetView(t *tab, p *structs.Page) {
	if p.Width == termW {
		// No changes to make
		return
	}
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

	oldFav := t.page.Favicon

	t.page = p

	go func() {
		parsed, _ := url.Parse(p.URL)
		handleFavicon(t, parsed.Host, oldFav)
	}()

	// Change page on screen
	t.view.SetText(p.Content)
	t.view.Highlight("") // Turn off highlights, other funcs may restore if necessary
	t.view.ScrollToBeginning()

	// Setup display
	App.SetFocus(t.view)

	// Save bottom bar for the tab - other funcs will apply/display it
	t.barLabel = ""
	t.barText = p.URL
}

// handleHTTP is used by handleURL.
// It opens HTTP links and displays Info and Error modals.
// Returns false if there was an error.
func handleHTTP(u string, showInfo bool) bool {
	if len(config.HTTPCommand) == 1 {
		// Possibly a non-command

		switch strings.TrimSpace(config.HTTPCommand[0]) {
		case "", "off":
			Error("HTTP Error", "Opening HTTP URLs is turned off.")
			return false
		case "default":
			s, err := webbrowser.Open(u)
			if err != nil {
				Error("Webbrowser Error", err.Error())
				return false
			}
			if showInfo {
				Info(s)
			}
			return true
		}
	}

	// Custom command
	var err error = nil
	if len(config.HTTPCommand) > 1 {
		err = exec.Command(config.HTTPCommand[0], append(config.HTTPCommand[1:], u)...).Start()
	} else {
		err = exec.Command(config.HTTPCommand[0], u).Start()
	}
	if err != nil {
		Error("HTTP Error", "Error executing custom browser command: "+err.Error())
		return false
	}

	App.Draw()
	return true
}

// handleOther is used by handleURL.
// It opens links other than Gemini and HTTP and displays Error modals.
func handleOther(u string) {
	// The URL should have a scheme due to a previous call to normalizeURL
	parsed, _ := url.Parse(u)

	// Search for a handler for the URL scheme
	handler := strings.TrimSpace(viper.GetString("url-handlers." + parsed.Scheme))
	if len(handler) == 0 {
		handler = strings.TrimSpace(viper.GetString("url-handlers.other"))
	}
	switch handler {
	case "", "off":
		Error("URL Error", "Opening "+parsed.Scheme+" URLs is turned off.")
	default:
		// The config has a custom command to execute for URLs
		fields := strings.Fields(handler)
		err := exec.Command(fields[0], append(fields[1:], u)...).Start()
		if err != nil {
			Error("URL Error", "Error executing custom command: "+err.Error())
		}
	}
	App.Draw()
}

// handleFavicon handles getting and displaying a favicon.
// `old` is the previous favicon for the tab.
func handleFavicon(t *tab, host, old string) {
	defer func() {
		// Update display if needed
		if t.page.Favicon != old && isValidTab(t) {
			rewriteTabRow()
		}
	}()

	if !viper.GetBool("a-general.emoji_favicons") {
		// Not enabled
		return
	}
	if t.page.Favicon != "" {
		return
	}
	if host == "" {
		return
	}

	fav := cache.GetFavicon(host)
	if fav == cache.KnownNoFavicon {
		// It's been cached that this host doesn't have a favicon
		return
	}
	if fav != "" {
		t.page.Favicon = fav
		rewriteTabRow()
		return
	}

	// No favicon cached
	res, err := client.Fetch("gemini://" + host + "/favicon.txt")
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		cache.AddFavicon(host, cache.KnownNoFavicon)
		return
	}
	defer res.Body.Close()

	if res.Status != 20 {
		cache.AddFavicon(host, cache.KnownNoFavicon)
		return
	}
	if !strings.HasPrefix(res.Meta, "text/") {
		cache.AddFavicon(host, cache.KnownNoFavicon)
		return
	}
	// It's a regular plain response

	buf := new(bytes.Buffer)
	_, err = io.CopyN(buf, res.Body, 29+2+1) // 29 is the max emoji length, +2 for CRLF, +1 so that the right size will EOF
	if err == nil {
		// Content was too large
		cache.AddFavicon(host, cache.KnownNoFavicon)
		return
	} else if err != io.EOF {
		// Some network reading error
		// No favicon is NOT known, could be a temporary error
		return
	}
	// EOF, which is what we want.
	emoji := strings.TrimRight(buf.String(), "\r\n")
	if !isemoji.IsEmoji(emoji) {
		cache.AddFavicon(host, cache.KnownNoFavicon)
		return
	}
	// Valid favicon found
	t.page.Favicon = emoji
	cache.AddFavicon(host, emoji)
}

// goURL is like handleURL, but takes care of history and the bottomBar.
// It should be preferred over handleURL in most cases.
// It has no return values to be processed.
//
// It should be called in a goroutine.
func goURL(t *tab, u string) {
	final, displayed := handleURL(t, u, 0)
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
//
// numRedirects is the number of redirects that resulted in the provided URL.
// It should typically be 0.
func handleURL(t *tab, u string, numRedirects int) (string, bool) {
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

		go func(p *structs.Page) {
			if b && t.hasContent() && !feeds.IsTracked(s) && viper.GetBool("feeds.popup") {
				// The current page might be an untracked feed, and the user wants
				// to be notified in such cases.

				feed, isFeed := getFeedFromPage(p)
				if isFeed && isValidTab(t) && t.page == p {
					// After parsing and track-checking time, the page is still being displayed
					addFeedDirect(p.URL, feed, false)
				}
			}
		}(t.page)

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
	if u == "about:feeds" {
		Feeds(t)
		return ret("about:feeds", true)
	}

	u = normalizeURL(u)
	u = cache.Redirect(u)

	parsed, err := url.Parse(u)
	if err != nil {
		Error("URL Error", err.Error())
		return ret("", false)
	}

	proxy := strings.TrimSpace(viper.GetString("proxies." + parsed.Scheme))
	usingProxy := false

	proxyHostname, proxyPort, err := net.SplitHostPort(proxy)
	if err != nil {
		// Error likely means there's no port in the host
		proxyHostname = proxy
		proxyPort = "1965"
	}

	if strings.HasPrefix(u, "http") {
		if proxy == "" || proxy == "off" {
			// No proxy available
			handleHTTP(u, true)
			return ret("", false)
		}
		usingProxy = true
	}

	if !strings.HasPrefix(u, "http") && !strings.HasPrefix(u, "gemini") {
		// Not a Gemini URL
		if proxy == "" || proxy == "off" {
			// No proxy available
			handleOther(u)
			return ret("", false)
		}
		usingProxy = true
	}

	// Gemini URL, or one with a Gemini proxy available

	// Load page from cache if it exists,
	// and this isn't a page that was redirected to by the server (indicates dynamic content)
	if numRedirects == 0 {
		page, ok := cache.GetPage(u)
		if ok {
			setPage(t, page)
			return ret(u, true)
		}
	}
	// Otherwise download it
	bottomBar.SetText("Loading...")
	t.barText = "Loading..." // Save it too, in case the tab switches during loading
	t.mode = tabModeLoading
	App.Draw()

	var res *gemini.Response
	if usingProxy {
		res, err = client.FetchWithProxy(proxyHostname, proxyPort, u)
	} else {
		res, err = client.Fetch(u)
	}

	// Loading may have taken a while, make sure tab is still valid
	if !isValidTab(t) {
		return ret("", false)
	}

	if errors.Is(err, client.ErrTofu) {
		if usingProxy {
			// They are using a proxy
			if Tofu(proxy, client.GetExpiry(proxyHostname, proxyPort)) {
				// They want to continue anyway
				client.ResetTofuEntry(proxyHostname, proxyPort, res.Cert)
				// Response can be used further down, no need to reload
			} else {
				// They don't want to continue
				return ret("", false)
			}
		} else {
			if Tofu(parsed.Host, client.GetExpiry(parsed.Hostname(), parsed.Port())) {
				// They want to continue anyway
				client.ResetTofuEntry(parsed.Hostname(), parsed.Port(), res.Cert)
				// Response can be used further down, no need to reload
			} else {
				// They don't want to continue
				return ret("", false)
			}
		}
	} else if err != nil {
		Error("URL Fetch Error", err.Error())
		return ret("", false)
	}
	if renderer.CanDisplay(res) {
		page, err := renderer.MakePage(u, res, textWidth(), leftMargin(), usingProxy)
		// Rendering may have taken a while, make sure tab is still valid
		if !isValidTab(t) {
			return ret("", false)
		}

		if errors.Is(err, renderer.ErrTooLarge) {
			// Make new request for downloading purposes
			res, clientErr := client.Fetch(u)
			if clientErr != nil && !errors.Is(clientErr, client.ErrTofu) {
				Error("URL Fetch Error", err.Error())
				return ret("", false)
			}
			go dlChoice("That page is too large. What would you like to do?", u, res)
			return ret("", false)
		}
		if errors.Is(err, renderer.ErrTimedOut) {
			// Make new request for downloading purposes
			res, clientErr := client.Fetch(u)
			if clientErr != nil && !errors.Is(clientErr, client.ErrTofu) {
				Error("URL Fetch Error", err.Error())
				return ret("", false)
			}
			go dlChoice("Loading that page timed out. What would you like to do?", u, res)
			return ret("", false)
		}
		if err != nil {
			Error("Page Error", "Issuing creating page: "+err.Error())
			return ret("", false)
		}

		page.Width = termW

		if !client.HasClientCert(parsed.Host) {
			// Don't cache pages with client certs
			go cache.AddPage(page)
		}

		setPage(t, page)
		return ret(u, true)
	}
	// Not displayable
	// Could be a non 20 (or 21) status code, or a different kind of document

	// Handle each status code
	switch res.Status {
	case 10, 11:
		userInput, ok := Input(res.Meta)
		if ok {
			// Make another request with the query string added
			// + chars are replaced because PathEscape doesn't do that
			parsed.RawQuery = gemini.QueryEscape(userInput)
			if len(parsed.String()) > gemini.URLMaxLength {
				Error("Input Error", "URL for that input would be too long.")
				return ret("", false)
			}
			return ret(handleURL(t, parsed.String(), 0))
		}
		return ret("", false)
	case 30, 31:
		parsedMeta, err := url.Parse(res.Meta)
		if err != nil {
			Error("Redirect Error", "Invalid URL: "+err.Error())
			return ret("", false)
		}
		redir := parsed.ResolveReference(parsedMeta).String()
		// Prompt before redirecting to non-Gemini protocol
		redirect := false
		if !strings.HasPrefix(redir, "gemini") {
			if YesNo("Follow redirect to non-Gemini URL?\n" + redir) {
				redirect = true
			} else {
				return ret("", false)
			}
		}
		// Prompt before redirecting
		autoRedirect := viper.GetBool("a-general.auto_redirect")
		if redirect || (autoRedirect && numRedirects < 5) || YesNo("Follow redirect?\n"+redir) {
			if res.Status == gemini.StatusRedirectPermanent {
				go cache.AddRedir(u, redir)
			}
			return ret(handleURL(t, redir, numRedirects+1))
		}
		return ret("", false)
	case 40:
		Error("Temporary Failure", escapeMeta(res.Meta))
		return ret("", false)
	case 41:
		Error("Server Unavailable", escapeMeta(res.Meta))
		return ret("", false)
	case 42:
		Error("CGI Error", escapeMeta(res.Meta))
		return ret("", false)
	case 43:
		Error("Proxy Failure", escapeMeta(res.Meta))
		return ret("", false)
	case 44:
		Error("Slow Down", "You should wait "+escapeMeta(res.Meta)+" seconds before making another request.")
		return ret("", false)
	case 50:
		Error("Permanent Failure", escapeMeta(res.Meta))
		return ret("", false)
	case 51:
		Error("Not Found", escapeMeta(res.Meta))
		return ret("", false)
	case 52:
		Error("Gone", escapeMeta(res.Meta))
		return ret("", false)
	case 53:
		Error("Proxy Request Refused", escapeMeta(res.Meta))
		return ret("", false)
	case 59:
		Error("Bad Request", escapeMeta(res.Meta))
		return ret("", false)
	case 60:
		Error("Client Certificate Required", escapeMeta(res.Meta))
		return ret("", false)
	case 61:
		Error("Certificate Not Authorised", escapeMeta(res.Meta))
		return ret("", false)
	case 62:
		Error("Certificate Not Valid", escapeMeta(res.Meta))
		return ret("", false)
	}

	// Status code 20, but not a document that can be displayed

	// First see if it's a feed, and ask the user about adding it if it is
	filename := path.Base(parsed.Path)
	mediatype, _, _ := mime.ParseMediaType(res.Meta)
	feed, ok := feeds.GetFeed(mediatype, filename, res.Body)
	if ok {
		go func() {
			added := addFeedDirect(u, feed, feeds.IsTracked(u))
			if !added {
				// Otherwise offer download choices
				go dlChoice("That file could not be displayed. What would you like to do?", u, res)
			}
		}()
		return ret("", false)
	}

	// Otherwise offer download choices
	go dlChoice("That file could not be displayed. What would you like to do?", u, res)
	return ret("", false)
}

// rewriteTabRow clears the tabRow and writes all the tabs number/favicons into it.
func rewriteTabRow() {
	tabRow.Clear()
	if viper.GetBool("a-general.color") {
		for i := 0; i < NumTabs(); i++ {
			char := strconv.Itoa(i + 1)
			if tabs[i].page.Favicon != "" {
				char = tabs[i].page.Favicon
			}
			fmt.Fprintf(tabRow, `["%d"][%s]  %s  [%s][""]|`,
				i,
				config.GetColorString("tab_num"),
				char,
				config.GetColorString("tab_divider"),
			)
		}
	} else {
		for i := 0; i < NumTabs(); i++ {
			char := strconv.Itoa(i + 1)
			if tabs[i].page.Favicon != "" {
				char = tabs[i].page.Favicon
			}
			fmt.Fprintf(tabRow, `["%d"]  %s  [""]|`, i, char)
		}
	}
	tabRow.Highlight(strconv.Itoa(curTab)).ScrollToHighlight()
	App.Draw()
}
