package display

import (
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
)

var helpCells = strings.TrimSpace(`
?|Bring up this help.
Esc|Leave the help
Arrow keys, h/j/k/l|Scroll and move a page.
Tab|Navigate to the next item in a popup.
Shift-Tab|Navigate to the previous item in a popup.
b, Alt-Left|Go back a page
f, Alt-Right|Go forward a page
g|Go to top of document
G|Go to bottom of document
spacebar|Open bar at the bottom - type a URL or link number
Enter|On a page this will start link highlighting. Press Tab and Shift-Tab to pick different links. Press enter again to go to one.
Shift-NUMBER|Go to a specific tab.
Shift-0, )|Go to the last tab.
Ctrl-H|Go home
Ctrl-T|New tab
Ctrl-W|Close tab. For now, only the right-most tab can be closed.
Ctrl-R, R|Reload a page, discarding the cached version.
Ctrl-B|View bookmarks
Ctrl-D|Add, change, or remove a bookmark for the current page.
q, Ctrl-Q|Quit
`)

var helpTable = cview.NewTable().
	SetSelectable(false, false).
	SetFixed(1, 2).
	SetBorders(true).
	SetBordersColor(tcell.ColorGray)

// Help displays the help and keybindings.
func Help() {
	helpTable.ScrollToBeginning()
	tabPages.SwitchToPage("help")
	App.Draw()
}
