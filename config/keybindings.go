package config

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/spf13/viper"
)

const (
	CmdInvalid = 0
	CmdLink1   = 1
	CmdLink2   = 2
	CmdLink3   = 3
	CmdLink4   = 4
	CmdLink5   = 5
	CmdLink6   = 6
	CmdLink7   = 7
	CmdLink8   = 8
	CmdLink9   = 9
	CmdLink0   = 10
	CmdTab1    = 11
	CmdTab2    = 12
	CmdTab3    = 13
	CmdTab4    = 14
	CmdTab5    = 15
	CmdTab6    = 16
	CmdTab7    = 17
	CmdTab8    = 18
	CmdTab9    = 19
	CmdTab0    = 20
	CmdBottom  = iota
	CmdEdit
	CmdHome
	CmdBookmarks
	CmdAddBookmark
	CmdSave
	CmdReload
	CmdBack
	CmdForward
	CmdPgup
	CmdPgdn
	CmdNewTab
	CmdCloseTab
	CmdNextTab
	CmdPrevTab
	CmdQuit
	CmdHelp
)

type KeyBinding struct {
	key tcell.Key
	mod tcell.ModMask
	r   rune
}

var bindings map[KeyBinding]int
var tcellKeys map[string]tcell.Key

func parseBinding(cmd int, binding string) {
	bslice := strings.Split(binding, ":")
	var k tcell.Key
	var m tcell.ModMask = 0
	var r rune = 0

	if len(bslice) > 1 && bslice[0] == "Alt" {
		m = tcell.ModAlt
		bslice = bslice[1:]
	}

	if len(bslice) > 1 && bslice[0] == "Rune" {
		k = tcell.KeyRune
		r = []rune(bslice[1])[0]
	} else {
		var ok bool
		k, ok = tcellKeys[bslice[0]]
		if !ok { // Bad keybinding!  Quietly ignore...
			return
		}
		if strings.HasPrefix(bslice[0], "Ctrl") {
			m += tcell.ModCtrl
		}
	}

	bindings[KeyBinding{k, m, r}] = cmd
}

func KeyInit() {
	configBindings := map[int]string{
		CmdLink1:       "keybindings.bind_link1",
		CmdLink2:       "keybindings.bind_link2",
		CmdLink3:       "keybindings.bind_link3",
		CmdLink4:       "keybindings.bind_link4",
		CmdLink5:       "keybindings.bind_link5",
		CmdLink6:       "keybindings.bind_link6",
		CmdLink7:       "keybindings.bind_link7",
		CmdLink8:       "keybindings.bind_link8",
		CmdLink9:       "keybindings.bind_link9",
		CmdLink0:       "keybindings.bind_link0",
		CmdTab1:        "keybindings.bind_tab1",
		CmdTab2:        "keybindings.bind_tab2",
		CmdTab3:        "keybindings.bind_tab3",
		CmdTab4:        "keybindings.bind_tab4",
		CmdTab5:        "keybindings.bind_tab5",
		CmdTab6:        "keybindings.bind_tab6",
		CmdTab7:        "keybindings.bind_tab7",
		CmdTab8:        "keybindings.bind_tab8",
		CmdTab9:        "keybindings.bind_tab9",
		CmdTab0:        "keybindings.bind_tab0",
		CmdBottom:      "keybindings.bind_bottom",
		CmdEdit:        "keybindings.bind_edit",
		CmdHome:        "keybindings.bind_home",
		CmdBookmarks:   "keybindings.bind_bookmarks",
		CmdAddBookmark: "keybindings.bind_add_bookmark",
		CmdSave:        "keybindings.bind_save",
		CmdReload:      "keybindings.bind_reload",
		CmdBack:        "keybindings.bind_back",
		CmdForward:     "keybindings.bind_forward",
		CmdPgup:        "keybindings.bind_pgup",
		CmdPgdn:        "keybindings.bind_pgdn",
		CmdNewTab:      "keybindings.bind_new_tab",
		CmdCloseTab:    "keybindings.bind_close_tab",
		CmdNextTab:     "keybindings.bind_next_tab",
		CmdPrevTab:     "keybindings.bind_prev_tab",
		CmdQuit:        "keybindings.bind_quit",
		CmdHelp:        "keybindings.bind_help",
	}
	tcellKeys = make(map[string]tcell.Key)
	bindings = make(map[KeyBinding]int)

	for k, kname := range tcell.KeyNames {
		tcellKeys[kname] = k
	}

	for c, allb := range configBindings {
		allb = viper.GetString(allb)
		for _, b := range strings.Split(allb, ",") {
			parseBinding(c, b)
		}
	}
}

func TranslateKeyEvent(e *tcell.EventKey) int {
	var ok bool
	var cmd int
	k := e.Key()
	if k == tcell.KeyRune {
		cmd, ok = bindings[KeyBinding{k, e.Modifiers(), e.Rune()}]
	} else { // Sometimes tcell sets e.Rune() on non-KeyRune events.
		cmd, ok = bindings[KeyBinding{k, e.Modifiers(), 0}]
	}
	if ok {
		return cmd
	}
	return CmdInvalid
}
