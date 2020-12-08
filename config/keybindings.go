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
	CmdSub
	CmdAddSub
)

type keyBinding struct {
	key tcell.Key
	mod tcell.ModMask
	r   rune
}

var bindings map[keyBinding]int
var tcellKeys map[string]tcell.Key

func parseBinding(cmd int, binding string) {
	var k tcell.Key
	var m tcell.ModMask = 0
	var r rune = 0

	if strings.HasPrefix(binding, "Alt-") {
		m = tcell.ModAlt
		binding = binding[4:]
	}

	if len(binding) == 1 {
		k = tcell.KeyRune
		r = []rune(binding)[0]
	} else if len(binding) == 0 {
		return
	} else if binding == "Space" {
		k = tcell.KeyRune
		r = ' '
	} else {
		var ok bool
		k, ok = tcellKeys[binding]
		if !ok { // Bad keybinding!  Quietly ignore...
			return
		}
		if strings.HasPrefix(binding, "Ctrl") {
			m += tcell.ModCtrl
		}
	}

	bindings[keyBinding{k, m, r}] = cmd
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
		CmdSub:         "keybindings.bind_sub",
		CmdAddSub:      "keybindings.bind_add_sub",
	}
	configTabNBindings := map[int]string{
		CmdTab1: "keybindings.bind_tab1",
		CmdTab2: "keybindings.bind_tab2",
		CmdTab3: "keybindings.bind_tab3",
		CmdTab4: "keybindings.bind_tab4",
		CmdTab5: "keybindings.bind_tab5",
		CmdTab6: "keybindings.bind_tab6",
		CmdTab7: "keybindings.bind_tab7",
		CmdTab8: "keybindings.bind_tab8",
		CmdTab9: "keybindings.bind_tab9",
		CmdTab0: "keybindings.bind_tab0",
	}
	tcellKeys = make(map[string]tcell.Key)
	bindings = make(map[keyBinding]int)

	for k, kname := range tcell.KeyNames {
		tcellKeys[kname] = k
	}

	for c, allb := range configBindings {
		for _, b := range viper.GetStringSlice(allb) {
			parseBinding(c, b)
		}
	}

	// Backwards compatibility with the old shift_numbers config line.
	shift_numbers := []rune(viper.GetString("keybindings.shift_numbers"))
	if len(shift_numbers) > 0 && len(shift_numbers) <= 10 {
		for i, r := range shift_numbers {
			bindings[keyBinding{tcell.KeyRune, 0, r}] = CmdTab1 + i
		}
	} else {
		for c, allb := range configTabNBindings {
			for _, b := range viper.GetStringSlice(allb) {
				parseBinding(c, b)
			}
		}
	}
}

func TranslateKeyEvent(e *tcell.EventKey) int {
	var ok bool
	var cmd int
	k := e.Key()
	if k == tcell.KeyRune {
		cmd, ok = bindings[keyBinding{k, e.Modifiers(), e.Rune()}]
	} else { // Sometimes tcell sets e.Rune() on non-KeyRune events.
		cmd, ok = bindings[keyBinding{k, e.Modifiers(), 0}]
	}
	if ok {
		return cmd
	}
	return CmdInvalid
}
