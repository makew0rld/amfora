package display

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/makeworld-the-better-one/amfora/config"
	"gitlab.com/tslocum/cview"
)

var helpCells = strings.TrimSpace(`
?|Bring up this help. You can scroll!
Esc|Leave the help
Arrow keys, h/j/k/l|Scroll and move a page.
%s|Go up a page in document
%s|Go down a page in document
g|Go to top of document
G|Go to bottom of document
Tab|Navigate to the next item in a popup.
Shift-Tab|Navigate to the previous item in a popup.
%s|Go back in the history
%s|Go forward in the history
%s|Open bar at the bottom - type a URL, link number, search term.
|You can also type two dots (..) to go up a directory in the URL.
|Typing new:N will open link number N in a new tab
|instead of the current one.
%s|Go to links 1-10 respectively.
%s|Edit current URL
Enter, Tab|On a page this will start link highlighting.
|Press Tab and Shift-Tab to pick different links.
|Press Enter again to go to one, or Esc to stop.
%s|Go to a specific tab. (Default: Shift-NUMBER)
%s|Go to the last tab.
%s|Previous tab
%s|Next tab
%s|Go home
%s|New tab, or if a link is selected,
|this will open the link in a new tab.
%s|Close tab. For now, only the right-most tab can be closed.
%s|Reload a page, discarding the cached version.
|This can also be used if you resize your terminal.
%s|View bookmarks
%s|Add, change, or remove a bookmark for the current page.
%s|Save the current page to your downloads.
%s|View subscriptions
%s|Add or update a subscription
%s|Quit
`)

var helpTable = cview.NewTable().
	SetSelectable(false, false).
	SetBorders(false).
	SetScrollBarVisibility(cview.ScrollBarNever)

// Help displays the help and keybindings.
func Help() {
	helpTable.ScrollToBeginning()
	tabPages.SwitchToPage("help")
	App.SetFocus(helpTable)
	App.Draw()
}

func helpInit() {
	// Populate help table
	helpTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc || key == tcell.KeyEnter {
			tabPages.SwitchToPage(strconv.Itoa(curTab))
			App.SetFocus(tabs[curTab].view)
			App.Draw()
		}
	})

	tabKeys := fmt.Sprintf("%s to %s", strings.Split(config.GetKeyBinding(config.CmdTab1), ",")[0],
		strings.Split(config.GetKeyBinding(config.CmdTab9), ",")[0])
	linkKeys := fmt.Sprintf("%s to %s", strings.Split(config.GetKeyBinding(config.CmdLink1), ",")[0],
		strings.Split(config.GetKeyBinding(config.CmdLink0), ",")[0])

	helpCells = fmt.Sprintf(helpCells,
		config.GetKeyBinding(config.CmdPgup),
		config.GetKeyBinding(config.CmdPgdn),
		config.GetKeyBinding(config.CmdBack),
		config.GetKeyBinding(config.CmdForward),
		config.GetKeyBinding(config.CmdBottom),
		linkKeys,
		config.GetKeyBinding(config.CmdEdit),
		tabKeys,
		config.GetKeyBinding(config.CmdTab0),
		config.GetKeyBinding(config.CmdPrevTab),
		config.GetKeyBinding(config.CmdNextTab),
		config.GetKeyBinding(config.CmdHome),
		config.GetKeyBinding(config.CmdNewTab),
		config.GetKeyBinding(config.CmdCloseTab),
		config.GetKeyBinding(config.CmdReload),
		config.GetKeyBinding(config.CmdBookmarks),
		config.GetKeyBinding(config.CmdAddBookmark),
		config.GetKeyBinding(config.CmdSave),
		config.GetKeyBinding(config.CmdSub),
		config.GetKeyBinding(config.CmdAddSub),
		config.GetKeyBinding(config.CmdQuit),
		)

	rows := strings.Count(helpCells, "\n") + 1
	cells := strings.Split(
		strings.ReplaceAll(helpCells, "\n", "|"),
		"|")
	cell := 0
	extraRows := 0 // Rows continued from the previous, without spacing
	for r := 0; r < rows; r++ {
		for c := 0; c < 2; c++ {
			var tableCell *cview.TableCell
			if c == 0 {
				// First column, the keybinding
				tableCell = cview.NewTableCell(" " + cells[cell]).
					SetAttributes(tcell.AttrBold).
					SetAlign(cview.AlignLeft)
			} else {
				tableCell = cview.NewTableCell(" " + cells[cell])
			}
			if c == 0 && cells[cell] == "" || (cell > 0 && cells[cell-1] == "" && c == 1) {
				// The keybinding column for this row was blank, meaning the explanation
				// column is continued from the previous row.
				// The row should be added without any spacing rows
				helpTable.SetCell(((2*r)-extraRows/2)-1, c, tableCell)
				extraRows++
			} else {
				helpTable.SetCell((2*r)-extraRows/2, c, tableCell) // Every other row, for readability
			}
			cell++
		}
	}
	tabPages.AddPage("help", helpTable, true, false)
}
