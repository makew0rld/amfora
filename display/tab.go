package display

import (
	"strings"
	"sync"

	"github.com/makeworld-the-better-one/amfora/structs"
	"gitlab.com/tslocum/cview"
)

type tabMode int

const (
	modeOff        tabMode = iota // Regular mode
	modeLinkSelect                // When the enter key is pressed, allow for tab-based link navigation
)

// tabHist holds the history for a tab.
type tabHistory struct {
	urls []string
	pos  int // Position: where in the list of URLs we are
}

// tab hold the information needed for each browser tab.
type tab struct {
	page        *structs.Page
	view        *cview.TextView
	mode        tabMode
	history     *tabHistory
	reformatMut *sync.Mutex // Mutex for reformatting, so there's only one reformat job at once
	selected    string      // The current text or link selected
	barLabel    string      // The bottomBar label for the tab
	barText     string      // The bottomBar text for the tab
}

// makeNewTab initializes an tab struct with no content.
func makeNewTab() *tab {
	return &tab{
		page: &structs.Page{},
		view: cview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			SetScrollable(true).
			SetWrap(false).
			SetChangedFunc(func() {
				App.Draw()
			}),
		mode:        modeOff,
		reformatMut: &sync.Mutex{},
	}
}

// addToHistory adds the given URL to history.
// It assumes the URL is currently being loaded and displayed on the page.
func (t *tab) addToHistory(u string) {
	if t.history.pos < len(t.history.urls)-1 {
		// We're somewhere in the middle of the history instead, with URLs ahead and behind.
		// The URLs ahead need to be removed so this new URL is the most recent item in the history
		t.history.urls = t.history.urls[:t.history.pos+1]
	}
	t.history.urls = append(t.history.urls, u)
	t.history.pos++
}

// pageUp scrolls up 75% of the height of the terminal, like Bombadillo.
func (t *tab) pageUp() {
	row, col := t.view.GetScrollOffset()
	t.view.ScrollTo(row-(termH/4)*3, col)
}

// pageDown scrolls down 75% of the height of the terminal, like Bombadillo.
func (t *tab) pageDown() {
	row, col := t.view.GetScrollOffset()
	t.view.ScrollTo(row+(termH/4)*3, col)
}

// hasContent returns true when the tab has a page that could be displayed.
// The most likely situation where false would be returned is when the default
// new tab content is being displayed.
func (t *tab) hasContent() bool {
	if t.page == nil || t.view == nil {
		return false
	}
	if t.page.Url == "" {
		return false
	}
	if strings.HasPrefix(t.page.Url, "about:") {
		return false
	}
	if t.page.Content == "" {
		return false
	}
	return true
}

// saveScroll saves where in the page the user was.
// It should be used whenever moving from one page to another.
func (t *tab) saveScroll() {
	// It will also be saved in the cache because the cache uses the same pointer
	row, col := t.view.GetScrollOffset()
	t.page.Row = row
	t.page.Column = col
}

// applyScroll applies the saved scroll values to the page and tab.
// It should only be used when going backward and forward.
func (t *tab) applyScroll() {
	t.view.ScrollTo(t.page.Row, t.page.Column)
}

// saveBottomBar saves the current bottomBar values in the tab.
func (t *tab) saveBottomBar() {
	t.barLabel = bottomBar.GetLabel()
	t.barText = bottomBar.GetText()
}

// applyBottomBar sets the bottomBar using the stored tab values
func (t *tab) applyBottomBar() {
	bottomBar.SetLabel(t.barLabel)
	bottomBar.SetText(t.barText)
}
