package display

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/structs"
	"gitlab.com/tslocum/cview"
)

type tabMode int

const (
	tabModeDone tabMode = iota
	tabModeLoading
)

type tabHistory struct {
	urls []string
	pos  int // Position: where in the list of URLs we are
}

// tab hold the information needed for each browser tab.
type tab struct {
	page     *structs.Page
	view     *cview.TextView
	history  *tabHistory
	mode     tabMode
	barLabel string // The bottomBar label for the tab
	barText  string // The bottomBar text for the tab
}

// makeNewTab initializes an tab struct with no content.
func makeNewTab() *tab {
	t := tab{
		page:    &structs.Page{Mode: structs.ModeOff},
		view:    cview.NewTextView(),
		history: &tabHistory{},
		mode:    tabModeDone,
	}
	t.view.SetDynamicColors(true)
	t.view.SetRegions(true)
	t.view.SetScrollable(true)
	t.view.SetWrap(false)
	t.view.SetScrollBarVisibility(config.ScrollBar)
	t.view.SetScrollBarColor(config.GetColor("scrollbar"))
	t.view.SetChangedFunc(func() {
		App.Draw()
	})
	t.view.SetDoneFunc(func(key tcell.Key) {
		// Altered from:
		// https://gitlab.com/tslocum/cview/-/blob/1f765c8695c3f4b35dae57f469d3aee0b1adbde7/demos/textview/main.go
		// Handles being able to select and "click" links with the enter and tab keys

		tab := curTab // Don't let it change in the middle of the code

		if tabs[tab].mode != tabModeDone {
			return
		}

		if key == tcell.KeyEsc {
			// Stop highlighting
			bottomBar.SetLabel("")
			bottomBar.SetText(tabs[tab].page.URL)
			tabs[tab].clearSelected()
			tabs[tab].saveBottomBar()
			return
		}

		if len(tabs[tab].page.Links) == 0 {
			// No links on page
			return
		}

		currentSelection := tabs[tab].view.GetHighlights()
		numSelections := len(tabs[tab].page.Links)

		if key == tcell.KeyEnter && len(currentSelection) > 0 {
			// A link is selected and enter was pressed: "click" it and load the page it's for
			bottomBar.SetLabel("")
			linkN, _ := strconv.Atoi(currentSelection[0])
			tabs[tab].page.Selected = tabs[tab].page.Links[linkN]
			tabs[tab].page.SelectedID = currentSelection[0]
			followLink(tabs[tab], tabs[tab].page.URL, tabs[tab].page.Links[linkN])
			return
		}
		if len(currentSelection) == 0 && (key == tcell.KeyEnter || key == tcell.KeyTab) {
			// They've started link highlighting
			tabs[tab].page.Mode = structs.ModeLinkSelect

			tabs[tab].view.Highlight("0")
			tabs[tab].view.ScrollToHighlight()
			// Display link URL in bottomBar
			bottomBar.SetLabel("[::b]Link: [::-]")
			bottomBar.SetText(tabs[tab].page.Links[0])
			tabs[tab].saveBottomBar()
			tabs[tab].page.Selected = tabs[tab].page.Links[0]
			tabs[tab].page.SelectedID = "0"
		}

		if len(currentSelection) > 0 {
			// There's still a selection, but a different key was pressed, not Enter

			index, _ := strconv.Atoi(currentSelection[0])
			if key == tcell.KeyTab {
				index = (index + 1) % numSelections
			} else if key == tcell.KeyBacktab {
				index = (index - 1 + numSelections) % numSelections
			} else {
				return
			}
			tabs[tab].view.Highlight(strconv.Itoa(index))
			tabs[tab].view.ScrollToHighlight()
			// Display link URL in bottomBar
			bottomBar.SetLabel("[::b]Link: [::-]")
			bottomBar.SetText(tabs[tab].page.Links[index])
			tabs[tab].saveBottomBar()
			tabs[tab].page.Selected = tabs[tab].page.Links[index]
			tabs[tab].page.SelectedID = strconv.Itoa(index)
		}
	})

	return &t
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

// hasContent returns false when the tab's page is malformed,
// has no content or URL, or if it's an 'about:' page.
func (t *tab) hasContent() bool {
	if t.page == nil || t.view == nil {
		return false
	}
	if t.page.URL == "" {
		return false
	}
	if strings.HasPrefix(t.page.URL, "about:") {
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

// clearSelected turns off any selection that was going on.
// It does not affect the bottomBar.
func (t *tab) clearSelected() {
	t.page.Mode = structs.ModeOff
	t.page.Selected = ""
	t.page.SelectedID = ""
	t.view.Highlight("")
}

// applySelected selects whatever is stored as the selected element in the struct,
// and sets the mode accordingly.
// It is safe to call if nothing was selected previously.
//
// applyBottomBar should be called after, as this func might set some bottomBar values.
func (t *tab) applySelected() {
	if t.page.Mode == structs.ModeOff {
		// Just in case
		t.page.Selected = ""
		t.page.SelectedID = ""
		t.view.Highlight("")
		return
	} else if t.page.Mode == structs.ModeLinkSelect {
		t.view.Highlight(t.page.SelectedID)

		if t.mode == tabModeDone {
			// Page is not loading so bottomBar can change
			t.barLabel = "[::b]Link: [::-]"
			t.barText = t.page.Selected
		}
	}
}

// applyAll uses applyScroll and applySelected to put a tab's TextView back the way it was.
// It also uses applyBottomBar if this is the current tab.
func (t *tab) applyAll() {
	t.applySelected()
	t.applyScroll()
	if t == tabs[curTab] {
		t.applyBottomBar()
	}
}
