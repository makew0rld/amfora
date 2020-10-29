package display

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/config"
	"gitlab.com/tslocum/cview"
)

var helpCells = strings.TrimSpace(`
?|Bring up this help. You can scroll!
Esc|Leave the help
Arrow keys, h/j/k/l|Scroll and move a page.
PgUp, u|Go up a page in document
PgDn, d|Go down a page in document
g|Go to top of document
G|Go to bottom of document
Tab|Navigate to the next item in a popup.
Shift-Tab|Navigate to the previous item in a popup.
b, Alt-Left|Go back in the history
f, Alt-Right|Go forward in the history
spacebar|Open bar at the bottom - type a URL, link number, search term.
|You can also type two dots (..) to go up a directory in the URL.
|Typing new:N will open link number N in a new tab
|instead of the current one.
Numbers|Go to links 1-10 respectively.
e|Edit current URL
Enter, Tab|On a page this will start link highlighting.
|Press Tab and Shift-Tab to pick different links.
|Press Enter again to go to one, or Esc to stop.
Shift-NUMBER|Go to a specific tab.
Shift-0, )|Go to the last tab.
F1|Previous tab
F2|Next tab
Ctrl-H|Go home
Ctrl-T|New tab, or if a link is selected,
|this will open the link in a new tab.
Ctrl-W|Close tab. For now, only the right-most tab can be closed.
Ctrl-R, R|Reload a page, discarding the cached version.
|This can also be used if you resize your terminal.
Ctrl-B|View bookmarks
Ctrl-D|Add, change, or remove a bookmark for the current page.
Ctrl-S|Save the current page to your downloads.
q, Ctrl-Q|Quit
Ctrl-C|Hard quit. This can be used when in the middle of downloading,
|for example.
`)

var helpTable = cview.NewTextView()

// Help displays the help and keybindings.
func Help() {
	helpTable.ScrollToBeginning()
	if !browser.HasTab("help") {
		browser.AddTab("help", "Help", helpTable)
	}
	browser.SetCurrentTab("help")
	App.SetFocus(helpTable)
	App.Draw()
}

func helpInit() {
	// Populate help table
	helpTable.SetBackgroundColor(config.GetColor("bg"))
	helpTable.SetPadding(0, 0, 1, 1)
	helpTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc || key == tcell.KeyEnter {
			browser.SetCurrentTab(strconv.Itoa(curTab))
			App.SetFocus(tabs[curTab].view)
			App.Draw()
		}
	})
	lines := strings.Split(helpCells, "\n")
	w := tabwriter.NewWriter(helpTable, 0, 8, 2, ' ', 0)
	for i, line := range lines {
		cells := strings.Split(line, "|")
		if i > 0 && len(cells[0]) > 0 {
			fmt.Fprintln(w, "\t")
		}
		fmt.Fprintf(w, "%s\t%s\n", cells[0], cells[1])
	}
	w.Flush()
	browser.AddTab("help", "Help", helpTable)
}
