package display

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/config"
	"gitlab.com/tslocum/cview"
)

var helpCells = strings.TrimSpace(
	"?\tBring up this help. You can scroll!\n" +
		"Esc\tLeave the help\n" +
		"Arrow keys, h/j/k/l\tScroll and move a page.\n" +
		"%s\tGo up a page in document\n" +
		"%s\tGo down a page in document\n" +
		"g\tGo to top of document\n" +
		"G\tGo to bottom of document\n" +
		"Tab\tNavigate to the next item in a popup.\n" +
		"Shift-Tab\tNavigate to the previous item in a popup.\n" +
		"%s\tGo back in the history\n" +
		"%s\tGo forward in the history\n" +
		"%s\tOpen bar at the bottom - type a URL, link number, search term.\n" +
		"\tYou can also type two dots (..) to go up a directory in the URL.\n" +
		"\tTyping new:N will open link number N in a new tab\n" +
		"\tinstead of the current one.\n" +
		"%s\tGo to links 1-10 respectively.\n" +
		"%s\tEdit current URL\n" +
		"Enter, Tab\tOn a page this will start link highlighting.\n" +
		"\tPress Tab and Shift-Tab to pick different links.\n" +
		"\tPress Enter again to go to one, or Esc to stop.\n" +
		"%s\tGo to a specific tab. (Default: Shift-NUMBER)\n" +
		"%s\tGo to the last tab.\n" +
		"%s\tPrevious tab\n" +
		"%s\tNext tab\n" +
		"%s\tGo home\n" +
		"%s\tNew tab, or if a link is selected,\n" +
		"\tthis will open the link in a new tab.\n" +
		"%s\tClose tab. For now, only the right-most tab can be closed.\n" +
		"%s\tReload a page, discarding the cached version.\n" +
		"\tThis can also be used if you resize your terminal.\n" +
		"%s\tView bookmarks\n" +
		"%s\tAdd, change, or remove a bookmark for the current page.\n" +
		"%s\tSave the current page to your downloads.\n" +
		"%s\tView subscriptions\n" +
		"%s\tAdd or update a subscription\n" +
		"%s\tQuit\n")

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
		if i > 0 && line[0] != '\t' {
			fmt.Fprintln(w, "\t")
		}
		fmt.Fprintln(w, line)
	}

	w.Flush()

	panels.AddPanel("help", helpTable, true, false)
}
