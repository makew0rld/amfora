package display

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/spf13/viper"
)

// This file contains the functions that aren't part of the public API.
// The funcs are for network and displaying.

// followLink should be used when the user "clicks" a link on a page.
// Not when a URL is opened on a new tab for the first time.
// It will handle setting the bottomBar.
func followLink(t *tab, prev, next string) {
	if strings.HasPrefix(next, "about:") {
		if final, ok := handleAbout(t, next); ok {
			t.addToHistory(final)
		}
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
		if strings.HasPrefix(p.URL, "gemini") ||
			strings.HasPrefix(p.URL, "about") ||
			strings.HasPrefix(p.URL, "file") {
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
