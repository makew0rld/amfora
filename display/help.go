package display

import (
	"fmt"
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

var helpTable = cview.NewTextView()

// Help displays the help and keybindings.
func Help() {
	helpTable.ScrollToBeginning()
	panels.ShowPanel("help")
	panels.SendToFront("help")
	App.SetFocus(helpTable)
}

func helpInit() {
	// Populate help table
	helpTable.SetBackgroundColor(config.GetColor("bg"))
	helpTable.SetPadding(0, 0, 1, 1)
	helpTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc || key == tcell.KeyEnter {
			panels.HidePanel("help")
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

	panels.AddPanel("help", helpTable, true, false)
}
