package display

import (
	"errors"
	"fmt"
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
	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/amfora/openssl"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/makeworld-the-better-one/amfora/subscriptions"
	"github.com/makeworld-the-better-one/amfora/sysopen"
	"github.com/makeworld-the-better-one/amfora/webbrowser"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/makeworld-the-better-one/rr"
	"github.com/spf13/viper"
)

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
	var proc *exec.Cmd
	if len(config.HTTPCommand) > 1 {
		proc = exec.Command(config.HTTPCommand[0], append(config.HTTPCommand[1:], u)...)
	} else {
		proc = exec.Command(config.HTTPCommand[0], u)
	}
	err := proc.Start()
	if err != nil {
		Error("HTTP Error", "Error executing custom browser command: "+err.Error())
		return false
	}
	//nolint:errcheck
	go proc.Wait() // Prevent zombies, see #219
	Info("Opened with: " + config.HTTPCommand[0])

	App.Draw()
	return true
}

// handleOther is used by handleURL.
// It opens links other than Gemini and HTTP and displays Error modals.
func handleOther(u string) {
	// The URL should have a scheme due to a previous call to normalizeURL
	parsed, _ := url.Parse(u)

	// Search for a handler for the URL scheme
	handler := viper.GetStringSlice("url-handlers." + parsed.Scheme)
	if len(handler) == 0 {
		// A string and not a list of strings, use old method of parsing
		// #214
		handler = strings.Fields(viper.GetString("url-handlers." + parsed.Scheme))
		if len(handler) == 0 {
			handler = viper.GetStringSlice("url-handlers.other")
			if len(handler) == 0 {
				handler = strings.Fields(viper.GetString("url-handlers.other"))
			}
		}
	}

	if len(handler) == 1 {
		// Maybe special key

		switch strings.TrimSpace(handler[0]) {
		case "", "off":
			Error("URL Error", "Opening "+parsed.Scheme+" URLs is turned off.")
			return
		case "default":
			_, err := sysopen.Open(u)
			if err != nil {
				Error("Application Error", err.Error())
				return
			}
			Info("Opened in default application")
			return
		}
	}

	// Custom application command

	var proc *exec.Cmd
	if len(handler) > 1 {
		proc = exec.Command(handler[0], append(handler[1:], u)...)
	} else {
		proc = exec.Command(handler[0], u)
	}
	err := proc.Start()
	if err != nil {
		Error("URL Error", "Error executing custom command: "+err.Error())
	}
	//nolint:errcheck
	go proc.Wait() // Prevent zombies, see #219
	Info("Opened with: " + handler[0])
	App.Draw()
}

// handleAbout can be called to deal with any URLs that start with
// 'about:'. It will display errors if the URL is not recognized,
// but not display anything if an 'about:' URL is not passed.
//
// It does not add the displayed page to history.
//
// It returns the URL displayed, and a bool indicating if the provided
// URL could be handled. The string returned will always be empty
// if the bool is false.
func handleAbout(t *tab, u string) (string, bool) {
	if !strings.HasPrefix(u, "about:") {
		return "", false
	}

	switch u {
	case "about:bookmarks":
		Bookmarks(t)
		return u, true
	case "about:newtab":
		temp := newTabPage // Copy
		setPage(t, &temp)
		t.applyBottomBar()
		return u, true
	case "about:version":
		temp := versionPage
		setPage(t, &temp)
		t.applyBottomBar()
		return u, true
	case "about:license":
		temp := licensePage
		setPage(t, &temp)
		t.applyBottomBar()
		return u, true
	case "about:thanks":
		temp := thanksPage
		setPage(t, &temp)
		t.applyBottomBar()
		return u, true
	case "about:about":
		temp := aboutPage
		setPage(t, &temp)
		t.applyBottomBar()
		return u, true
	}

	if u == "about:subscriptions" || (len(u) > 20 && u[:20] == "about:subscriptions?") {
		// about:subscriptions?2 views page 2
		return Subscriptions(t, u), true
	}
	if u == "about:manage-subscriptions" || (len(u) > 27 && u[:27] == "about:manage-subscriptions?") {
		ManageSubscriptions(t, u)
		// Don't count remove command in history
		if u == "about:manage-subscriptions" {
			return u, true
		}
		return "", false
	}

	Error("Error", "Not a valid 'about:' URL.")
	return "", false
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

		t.preferURLHandler = false

		go func(p *structs.Page) {
			if b && t.hasContent() && !t.isAnAboutPage() && viper.GetBool("subscriptions.popup") {
				// The current page might be an untracked feed, and the user wants
				// to be notified in such cases.

				feed, isFeed := getFeedFromPage(p)
				if isFeed && isValidTab(t) && t.page == p {
					// After parsing and track-checking time, the page is still being displayed
					addFeedDirect(p.URL, feed, subscriptions.IsSubscribed(p.URL))
				}
			}
		}(t.page)

		return s, b
	}

	t.barLabel = ""
	bottomBar.SetLabel("")

	App.SetFocus(t.view)

	if strings.HasPrefix(u, "about:") {
		return ret(handleAbout(t, u))
	}

	u = client.NormalizeURL(u)
	u = cache.Redirect(u)

	parsed, err := url.Parse(u)
	if err != nil {
		Error("URL Error", err.Error())
		return ret("", false)
	}

	// check if a prompt is needed to handle this url
	prompt := viper.GetBool("url-prompts.other")
	if viper.IsSet("url-prompts." + parsed.Scheme) {
		prompt = viper.GetBool("url-prompts." + parsed.Scheme)
	}
	if prompt && !(YesNo("Follow URL?\n" + u)) {
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
		if proxy == "" || proxy == "off" || t.preferURLHandler {
			// No proxy available
			handleHTTP(u, true)
			return ret("", false)
		}
		usingProxy = true
	}

	if strings.HasPrefix(u, "file") {
		page, ok := handleFile(u)
		if !ok {
			return ret("", false)
		}
		setPage(t, page)
		return ret(u, true)
	}

	if !strings.HasPrefix(u, "http") && !strings.HasPrefix(u, "gemini") && !strings.HasPrefix(u, "file") {
		// Not a Gemini URL
		if proxy == "" || proxy == "off" || t.preferURLHandler {
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

	// Fetch happened successfully, use RestartReader to buffer read data
	res.Body = rr.NewRestartReader(res.Body)

	if renderer.CanDisplay(res) {
		page, err := renderer.MakePage(u, res, textWidth(), usingProxy)
		// Rendering may have taken a while, make sure tab is still valid
		if !isValidTab(t) {
			return ret("", false)
		}

		if errors.Is(err, renderer.ErrTooLarge) {
			// Downloading now
			// Disable read timeout and go back to start
			res.SetReadTimeout(0) //nolint: errcheck
			res.Body.(*rr.RestartReader).Restart()
			dlChoice("That page is too large. What would you like to do?", u, res)
			return ret("", false)
		}
		if errors.Is(err, renderer.ErrTimedOut) {
			// Downloading now
			// Disable read timeout and go back to start
			res.SetReadTimeout(0) //nolint: errcheck
			res.Body.(*rr.RestartReader).Restart()
			dlChoice("Loading that page timed out. What would you like to do?", u, res)
			return ret("", false)
		}
		if err != nil {
			Error("Page Error", "Issuing creating page: "+err.Error())
			return ret("", false)
		}

		page.TermWidth = termW

		if !client.HasClientCert(parsed.Host, parsed.Path) {
			// Don't cache pages with client certs
			go cache.AddPage(page)
		}

		setPage(t, page)
		return ret(u, true)
	}
	// Not displayable
	// Could be a non 20 status code, or a different kind of document

	// Handle each status code
	// Except 20, that's handled after the switch
	status := gemini.CleanStatus(res.Status)
	switch status {
	case 10, 11:
		var userInput string
		var ok bool

		if status == 10 {
			// Regular input
			userInput, ok = Input(res.Meta, false)
		} else {
			// Sensitive input
			userInput, ok = Input(res.Meta, true)
		}
		if ok {
			// Make another request with the query string added
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
		justAddsSlash := (redir == u+"/")
		// Prompt before redirecting to non-Gemini protocol
		redirect := false
		if !justAddsSlash && !strings.HasPrefix(redir, "gemini") {
			if YesNo("Follow redirect to non-Gemini URL?\n" + redir) {
				redirect = true
			} else {
				return ret("", false)
			}
		}
		// Prompt before redirecting
		autoRedirect := justAddsSlash || viper.GetBool("a-general.auto_redirect")
		if redirect || (autoRedirect && numRedirects < 5) || YesNo("Follow redirect?\n"+redir) {
			if status == gemini.StatusRedirectPermanent {
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
		ok := YesNo("This page requires certificate which we could not find in your config.\nWould you like Amfora to create a new certificate for that page?")
		if !ok {
			Error("Client Certificate Required", escapeMeta(res.Meta))
			return ret("", false)
		}

		username, ok := Input("Certificate Username", false)
		if !ok {
			return ret("", false)
		}

		days, ok := Input("Certificate Expire Days (default: 1825)", false)
		if !ok {
			return ret("", false)
		}

		daysInt, err := strconv.ParseInt(days, 0, 0)
		if err != nil {
			return ret("Days supposed to be int value", false)
		}

		err = openssl.CallOpenSSL(t.page.URL, username, int(daysInt))
		if err != nil {
			logger.Logger.Fatal(err)
			return ret(err.Error(), true)
		}

		err = client.CreateNewCertRow(t.page.URL)
		if err != nil {
			logger.Logger.Fatal(err)
			return ret("", false)
		}

		Info("Creating certificate successfull")
		return ret("", false)
	case 61:
		Error("Certificate Not Authorised", escapeMeta(res.Meta))
		return ret("", false)
	case 62:
		Error("Certificate Not Valid", escapeMeta(res.Meta))
		return ret("", false)
	default:
		if !gemini.StatusInRange(status) {
			// Status code not in a valid range
			Error("Status Code Error", fmt.Sprintf("Out of range status code: %d", status))
			return ret("", false)
		}
	}

	// Status code 20, but not a document that can be displayed

	// First see if it's a feed, and ask the user about adding it if it is
	filename := path.Base(parsed.Path)
	mediatype, _, _ := mime.ParseMediaType(res.Meta)
	feed, ok := subscriptions.GetFeed(mediatype, filename, res.Body)
	if ok {
		go func() {
			added := addFeedDirect(u, feed, subscriptions.IsSubscribed(u))
			if !added {
				// Otherwise offer download choices
				// Disable read timeout and go back to start
				res.SetReadTimeout(0) //nolint: errcheck
				res.Body.(*rr.RestartReader).Restart()
				dlChoice("That file could not be displayed. What would you like to do?", u, res)
			}
		}()
		return ret("", false)
	}

	// Otherwise offer download choices
	// Disable read timeout and go back to start
	res.SetReadTimeout(0) //nolint: errcheck
	res.Body.(*rr.RestartReader).Restart()
	dlChoice("That file could not be displayed. What would you like to do?", u, res)
	return ret("", false)
}
