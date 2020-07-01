package display

import (
	"strconv"
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
b, Alt-Left|Go back in the history
f, Alt-Right|Go forward in the history
g|Go to top of document
G|Go to bottom of document
spacebar|Open bar at the bottom - type a URL, link number, or search term. You can also type two dots (..) to go up a directory in the URL, as well as new:N to open link number N in a new tab instead of the current one.
Enter|On a page this will start link highlighting. Press Tab and Shift-Tab to pick different links. Press Enter again to go to one, or Esc to stop.
Shift-NUMBER|Go to a specific tab.
Shift-0, )|Go to the last tab.
Ctrl-H|Go home
Ctrl-T|New tab, or if a link is selected, this will open the link in a new tab.
Ctrl-W|Close tab. For now, only the right-most tab can be closed.
Ctrl-R, R|Reload a page, discarding the cached version.
Ctrl-B|View bookmarks
Ctrl-D|Add, change, or remove a bookmark for the current page.
q, Ctrl-Q, Ctrl-C|Quit
`)

var helpTable = cview.NewTable().
	SetSelectable(false, false).
	SetBorders(true).
	SetBordersColor(tcell.ColorGray)

// Help displays the help and keybindings.
func Help() {
	helpTable.ScrollToBeginning()
	tabPages.SwitchToPage("help")
	App.Draw()
}

func helpInit() {
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
					SetAlign(cview.AlignCenter)
			} else {
				tableCell = cview.NewTableCell("  " + cells[cell])
			}
			helpTable.SetCell(r, c, tableCell)
			cell++
		}
	}
	tabPages.AddPage("help", helpTable, true, false)
}

// TODO: Wrap cell text so it's not offscreen
