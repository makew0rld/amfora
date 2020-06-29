package display

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/makeworld-the-better-one/amfora/cache"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

var curTab = -1                          // What number tab is currently visible, -1 means there are no tabs at all
var tabMap = make(map[int]*structs.Page) // Map of tab number to page
// Holds the actual tab primitives
var tabViews = make(map[int]*cview.TextView)

var termW int

// The user input and URL display bar at the bottom
var bottomBar = cview.NewInputField().
	SetFieldBackgroundColor(tcell.ColorWhite).
	SetFieldTextColor(tcell.ColorBlack)

// Viewer for the tab primitives
// Pages are named as strings of tab numbers - so the textview for the first tab
// is held in the page named "0".
// The only pages that don't confine to this scheme are those named after modals,
// which are used to draw modals on top the current tab.
// Ex: "info", "error", "input", "yesno"
var tabPages = cview.NewPages()

// The tabs at the top with titles
var tabRow = cview.NewTextView().
	SetDynamicColors(true).
	SetRegions(true).
	SetScrollable(true).
	SetWrap(false).
	SetHighlightedFunc(func(added, removed, remaining []string) {
		// There will always only be one string in added - never multiple highlights
		// Remaining should always be empty
		i, _ := strconv.Atoi(added[0])
		tabPages.SwitchToPage(strconv.Itoa(i)) // Tab names are just numbers, zero-indexed
	})

// Root layout
var layout = cview.NewFlex().
	SetDirection(cview.FlexRow).
	AddItem(tabRow, 1, 1, false).
	AddItem(nil, 1, 1, false). // One line of empty space above the page
	AddItem(tabPages, 0, 1, true).
	// AddItem(cview.NewFlex(). // The page text in the middle is held in another flex, to center it
	// 				SetDirection(cview.FlexColumn).
	// 				AddItem(nil, 0, 1, false).
	// 				AddItem(tabPages, 0, 7, true). // The text occupies 7/9 of the screen horizontally
	// 				AddItem(nil, 0, 1, false),
	// 				0, 1, true).
	AddItem(nil, 1, 1, false). // One line of empty space before bottomBar
	AddItem(bottomBar, 1, 1, false)

var App = cview.NewApplication().
	EnableMouse(false).
	SetRoot(layout, true).
	SetAfterResizeFunc(func(width int, height int) {
		// Store for calculations
		termW = width

		// Shift new tabs created before app startup, when termW == 0
		// XXX: This is hacky but works. The biggest issue is that there will sometimes be a tiny flash
		// of the old not shifted tab on startup.
		if tabMap[curTab] == &newTabPage {
			setLeftMargin(tabMap[curTab])
			tabViews[curTab].SetText(tabMap[curTab].Content)
		}
	})

var renderedNewTabContent string
var newTabLinks []string
var newTabPage structs.Page

func Init() {
	tabRow.SetChangedFunc(func() {
		App.Draw()
	})

	// Populate help table
	helpTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			tabPages.SwitchToPage(strconv.Itoa(curTab))
		}
	})
	rows := strings.Count(helpCells, "\n") + 1
	cells := strings.Split(
		strings.ReplaceAll(helpCells, "\n", "|"),
		"|")
	cell := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < 2; c++ {
			var tableCell *cview.TableCell
			if c == 0 {
				tableCell = cview.NewTableCell(cells[cell]).
					SetAttributes(tcell.AttrBold).
					SetExpansion(1).
					SetAlign(cview.AlignCenter)
			} else {
				tableCell = cview.NewTableCell("  " + cells[cell]).
					SetExpansion(2)
			}
			helpTable.SetCell(r, c, tableCell)
			cell++
		}
	}
	tabPages.AddPage("help", helpTable, true, false)

	if viper.GetBool("a-general.color") {
		bottomBar.SetLabelColor(tcell.ColorGreen)
	} else {
		bottomBar.SetLabelColor(tcell.ColorBlack)
	}
	bottomBar.SetBackgroundColor(tcell.ColorWhite)
	bottomBar.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// Figure out whether it's a URL, link number, or search
			// And send out a request

			query := bottomBar.GetText()

			if strings.TrimSpace(query) == "" {
				// Ignore
				bottomBar.SetLabel("")
				bottomBar.SetText(tabMap[curTab].Url)
				App.SetFocus(tabViews[curTab])
				return
			}

			i, err := strconv.Atoi(query)
			if err != nil {
				// It's a full URL or search term
				// Detect if it's a search or URL
				if strings.Contains(query, " ") || (!strings.Contains(query, "//") && !strings.Contains(query, ".") && !strings.HasPrefix(query, "about:")) {
					URL(viper.GetString("a-general.search") + "?" + pathEscape(query))
				} else {
					// Full URL
					URL(query)
				}
				bottomBar.SetLabel("")
				return
			}
			if i <= len(tabMap[curTab].Links) && i > 0 {
				// It's a valid link number
				followLink(tabMap[curTab].Url, tabMap[curTab].Links[i-1])
				bottomBar.SetLabel("")
				return
			}
			// Invalid link number
			bottomBar.SetLabel("")
			bottomBar.SetText(tabMap[curTab].Url)
			App.SetFocus(tabViews[curTab])

		case tcell.KeyEscape:
			// Set back to what it was
			bottomBar.SetLabel("")
			bottomBar.SetText(tabMap[curTab].Url)
			App.SetFocus(tabViews[curTab])
		}
		// Other potential keys are Tab and Backtab, they are ignored
	})

	// Render the default new tab content ONCE and store it for later
	renderedNewTabContent, newTabLinks = renderer.RenderGemini(newTabContent, textWidth())
	newTabPage = structs.Page{Content: renderedNewTabContent, Links: newTabLinks, Url: "about:newtab"}

	modalInit()

	// Setup map of keys to functions here
	// Changing tabs, new tab, etc
	App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		_, ok := App.GetFocus().(*cview.Button)
		if ok {
			// It's focused on a modal right now, nothing should interrupt
			return event
		}
		_, ok = App.GetFocus().(*cview.InputField)
		if ok {
			// An InputField is in focus, nothing should interrupt
			return event
		}

		// History arrow keys
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyLeft {
				histBack()
				return nil
			}
			if event.Key() == tcell.KeyRight {
				histForward()
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlT:
			NewTab()
			return nil
		case tcell.KeyCtrlW:
			CloseTab()
			return nil
		case tcell.KeyCtrlR:
			Reload()
			return nil
		case tcell.KeyCtrlH:
			URL(viper.GetString("a-general.home"))
			return nil
		case tcell.KeyCtrlQ:
			Stop()
			return nil
		case tcell.KeyCtrlB:
			Bookmarks()
			addToHist("about:bookmarks")
			return nil
		case tcell.KeyCtrlD:
			go addBookmark()
			return nil
		case tcell.KeyRune:
			// Regular key was sent
			switch string(event.Rune()) {
			case " ":
				// Space starts typing, like Bombadillo
				bottomBar.SetLabel("[::b]URL/Num./Search: [::-]")
				bottomBar.SetText("")
				App.SetFocus(bottomBar)
				return nil
			case "q":
				Stop()
				return nil
			case "R":
				Reload()
				return nil
			case "b":
				histBack()
				return nil
			case "f":
				histForward()
				return nil
			case "?":
				Help()
				return nil
			// Shift+NUMBER keys, for switching to a specific tab
			case "!":
				SwitchTab(0)
				return nil
			case "@":
				SwitchTab(1)
				return nil
			case "#":
				SwitchTab(2)
				return nil
			case "$":
				SwitchTab(3)
				return nil
			case "%":
				SwitchTab(4)
				return nil
			case "^":
				SwitchTab(5)
				return nil
			case "&":
				SwitchTab(6)
				return nil
			case "*":
				SwitchTab(7)
				return nil
			case "(":
				SwitchTab(8)
				return nil
			case ")": // Zero key goes to the last tab
				SwitchTab(NumTabs() - 1)
				return nil
			}
		}
		return event
	})
}

// Stop stops the app gracefully.
// In the future it will handle things like ongoing downloads, etc
func Stop() {
	App.Stop()
}

// NewTab opens a new tab and switches to it, displaying the
// the default empty content because there's no URL.
func NewTab() {
	// Create TextView in tabViews and change curTab
	// Set the textView options, and the changed func to App.Draw()
	// SetDoneFunc to do link highlighting
	// Add view to pages and switch to it

	curTab = NumTabs()
	tabMap[curTab] = &newTabPage
	setLeftMargin(tabMap[curTab])
	tabViews[curTab] = cview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		SetWrap(false).
		SetText(tabMap[curTab].Content).
		ScrollToBeginning().
		SetChangedFunc(func() {
			App.Draw()
		}).
		SetDoneFunc(func(key tcell.Key) {
			// Altered from: https://gitlab.com/tslocum/cview/-/blob/master/demos/textview/main.go
			// Handles being able to select and "click" links with the enter and tab keys

			currentSelection := tabViews[curTab].GetHighlights()
			numSelections := len(tabMap[curTab].Links)
			if key == tcell.KeyEnter {
				if len(currentSelection) > 0 && len(tabMap[curTab].Links) > 0 {
					// A link was selected, "click" it and load the page it's for
					linkN, _ := strconv.Atoi(currentSelection[0])
					followLink(tabMap[curTab].Url, tabMap[curTab].Links[linkN])
					return
				} else {
					tabViews[curTab].Highlight("0").ScrollToHighlight()
				}
			} else if len(currentSelection) > 0 {
				index, _ := strconv.Atoi(currentSelection[0])
				if key == tcell.KeyTab {
					index = (index + 1) % numSelections
				} else if key == tcell.KeyBacktab {
					index = (index - 1 + numSelections) % numSelections
				} else {
					return
				}
				tabViews[curTab].Highlight(strconv.Itoa(index)).ScrollToHighlight()
			}
		})

	tabHist[curTab] = []string{}
	// Can't go backwards, but this isn't the first page either.
	// The first page will be the next one the user goes to.
	tabHistPos[curTab] = -1

	tabPages.AddAndSwitchToPage(strconv.Itoa(curTab), tabViews[curTab], true)
	App.SetFocus(tabViews[curTab])

	// Add tab number to the actual place where tabs are show on the screen
	// Tab regions are 0-indexed but text displayed on the screen starts at 1
	if viper.GetBool("a-general.color") {
		fmt.Fprintf(tabRow, `["%d"][darkcyan]  %d  [white][""]|`, curTab, curTab+1)
	} else {
		fmt.Fprintf(tabRow, `["%d"]  %d  [""]|`, curTab, curTab+1)
	}
	tabRow.Highlight(strconv.Itoa(curTab)).ScrollToHighlight()

	bottomBar.SetLabel("")
	bottomBar.SetText("")

	// Draw just in case
	App.Draw()
}

// CloseTab closes the current tab and switches to the one to its left.
func CloseTab() {
	// Basically the NewTab() func inverted

	// TODO: Support closing middle tabs, by renumbering all the maps
	// So that tabs to the right of the closed tabs point to the right places
	// For now you can only close the right-most tab
	if curTab != NumTabs()-1 {
		return
	}

	if NumTabs() <= 1 {
		// There's only one tab open, close the app instead
		Stop()
		return
	}

	delete(tabMap, curTab)
	tabPages.RemovePage(strconv.Itoa(curTab))
	delete(tabViews, curTab)

	delete(tabHist, curTab)
	delete(tabHistPos, curTab)

	if curTab <= 0 {
		curTab = NumTabs() - 1
	} else {
		curTab--
	}

	tabPages.SwitchToPage(strconv.Itoa(curTab)) // Go to previous page
	// Rewrite the tab display
	tabRow.Clear()
	if viper.GetBool("a-general.color") {
		for i := 0; i < NumTabs(); i++ {
			fmt.Fprintf(tabRow, `["%d"][darkcyan]  %d  [white][""]|`, i, i+1)
		}
	} else {
		for i := 0; i < NumTabs(); i++ {
			fmt.Fprintf(tabRow, `["%d"]  %d  [""]|`, i, i+1)
		}
	}
	tabRow.Highlight(strconv.Itoa(curTab)).ScrollToHighlight()

	bottomBar.SetLabel("")
	bottomBar.SetText(tabMap[curTab].Url)

	// Just in case
	App.Draw()
}

// SwitchTab switches to a specific tab, using its number, 0-indexed.
// The tab numbers are clamped to the end, so for example numbers like -5 and 1000 are still valid.
// This means that calling something like SwitchTab(curTab - 1) will never cause an error.
func SwitchTab(tab int) {
	if tab < 0 {
		tab = 0
	}
	if tab > NumTabs()-1 {
		tab = NumTabs() - 1
	}

	curTab = tab % NumTabs()
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	tabRow.Highlight(strconv.Itoa(curTab)).ScrollToHighlight()

	bottomBar.SetLabel("")
	bottomBar.SetText(tabMap[curTab].Url)

	// Just in case
	App.Draw()
}

func Reload() {
	cache.Remove(tabMap[curTab].Url)
	tabMap[curTab].LeftMargin = 0 // Redo left margin
	go handleURL(tabMap[curTab].Url)
}

// URL loads and handles the provided URL for the current tab.
// It should be an absolute URL.
func URL(u string) {
	// Some code is copied in followLink()

	if u == "about:bookmarks" {
		Bookmarks()
		addToHist("about:bookmarks")
		return
	}
	if u == "about:newtab" {
		setPage(&newTabPage)
		return
	}
	if strings.HasPrefix(u, "about:") {
		Error("Error", "Not a valid 'about:' URL.")
		return
	}

	go func() {
		final, displayed := handleURL(u)
		if displayed {
			addToHist(final)
		}
	}()
}

func NumTabs() int {
	return len(tabViews)
}
