package display

// This file contains the functions that aren't part of the public API.
// The funcs are for network and displaying.

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
)

// followLink should be used when the user "clicks" a link on a page,
// but not when a URL is opened on a new tab for the first time.
//
// It will handle updating the bottomBar.
//
// It should be called with the `go` keyword to spawn a new goroutine if
// it would otherwise block the UI loop, such as when called from an input
// handler.
//
// It blocks until navigation is finished, and we've completed any user
// interaction related to loading the URL (such as info, error modals)
func followLink(t *tab, prev, next string) {
	if strings.HasPrefix(next, "about:") {
		goURL(t, next)
		return
	}

	if t.hasContent() && !t.isAnAboutPage() {
		nextURL, err := resolveRelLink(t, prev, next)
		if err != nil {
			Error("URL Error", err.Error())
			return
		}
		goURL(t, nextURL)
		return
	}
	// No content on current tab, so the "prev" URL is not valid.
	// An example is the about:newtab page
	_, err := url.Parse(next)
	if err != nil {
		Error("URL Error", "Link URL could not be parsed")
		return
	}
	goURL(t, next)
}

// reformatPage will take the raw page content and reformat it according to the current terminal dimensions.
// It should be called when the terminal size changes.
// It will not waste resources if the passed page is already fitted to the current terminal width, and can be
// called safely even when the page might be already formatted properly.
func reformatPage(p *structs.Page) {
	if p.TermWidth == termW {
		// No changes to make
		return
	}

	// TODO: Setup a renderer.RenderFromMediatype func so this isn't needed

	var rendered string
	switch p.Mediatype {
	case structs.TextGemini:
		// Links are not recorded because they won't change
		proxied := true
		if strings.HasPrefix(p.URL, "gemini") ||
			strings.HasPrefix(p.URL, "about") ||
			strings.HasPrefix(p.URL, "file") {
			proxied = false
		}
		rendered, _ = renderer.RenderGemini(p.Raw, textWidth(), proxied)
	case structs.TextPlain:
		rendered = renderer.RenderPlainText(p.Raw)
	case structs.TextAnsi:
		rendered = renderer.RenderANSI(p.Raw)
	default:
		// Rendering this type is not implemented
		return
	}
	p.Content = rendered
	p.TermWidth = termW
}

// reformatPageAndSetView is for reformatting a page that is already being displayed.
// setPage should be used when a page is being loaded for the first time.
func reformatPageAndSetView(t *tab, p *structs.Page) {
	if p.TermWidth == termW {
		// No changes to make
		return
	}
	reformatPage(p)
	t.view.SetText(p.Content)
	t.applyScroll() // Go back to where you were, roughly

	App.Draw()
}

// setPage displays a Page on the passed tab number.
// The bottomBar is not actually changed in this func
func setPage(t *tab, p *structs.Page) {
	if !isValidTab(t) {
		// Don't waste time reformatting an invalid tab
		return
	}

	// Make sure the page content is fitted to the terminal every time it's displayed
	reformatPage(p)

	t.page = p

	// Change page on screen
	t.view.SetText(p.Content)
	t.view.Highlight("") // Turn off highlights, other funcs may restore if necessary
	t.view.ScrollToBeginning()
	// Reset page left margin
	tabNum := tabNumber(t)
	browser.AddTab(
		strconv.Itoa(tabNum),
		makeTabLabel(strconv.Itoa(tabNum+1)),
		makeContentLayout(t.view, leftMargin()),
	)
	App.Draw()

	// Setup display
	App.SetFocus(t.view)

	// Save bottom bar for the tab - other funcs will apply/display it
	t.barLabel = ""
	t.barText = p.URL
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
	} else if t.page.URL == "" {
		// The tab is showing interstitial or no content. Let's go to about:newtab.
		handleAbout(t, "about:newtab")
	}
	if t == tabs[curTab] {
		// Display the bottomBar state that handleURL set
		t.applyBottomBar()
	}
}
