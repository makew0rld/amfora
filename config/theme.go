package config

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Functions to allow themeing configuration.
// UI element colors are mapped to a string key, such as "error" or "tab_bg"
// These are the same keys used in the config file.

var themeMu = sync.RWMutex{}
var theme = map[string]tcell.Color{
	// Default values below

	"bg":              tcell.ColorBlack, // Used for cview.Styles.PrimitiveBackgroundColor
	"tab_num":         tcell.Color30,    // xterm:Turquoise4, #008787
	"tab_divider":     tcell.ColorWhite,
	"bottombar_label": tcell.Color30,
	"bottombar_text":  tcell.ColorBlack,
	"bottombar_bg":    tcell.ColorWhite,
	"scrollbar":       tcell.ColorWhite,

	// Modals
	"btn_bg":   tcell.ColorNavy, // All modal buttons
	"btn_text": tcell.ColorWhite,

	"dl_choice_modal_bg":      tcell.ColorPurple,
	"dl_choice_modal_text":    tcell.ColorWhite,
	"dl_modal_bg":             tcell.Color130, // xterm:DarkOrange3, #af5f00
	"dl_modal_text":           tcell.ColorWhite,
	"info_modal_bg":           tcell.ColorGray,
	"info_modal_text":         tcell.ColorWhite,
	"error_modal_bg":          tcell.ColorMaroon,
	"error_modal_text":        tcell.ColorWhite,
	"yesno_modal_bg":          tcell.ColorPurple,
	"yesno_modal_text":        tcell.ColorWhite,
	"tofu_modal_bg":           tcell.ColorMaroon,
	"tofu_modal_text":         tcell.ColorWhite,
	"subscription_modal_bg":   tcell.Color61, // xterm:SlateBlue3, #5f5faf
	"subscription_modal_text": tcell.ColorWhite,

	"input_modal_bg":         tcell.ColorGreen,
	"input_modal_text":       tcell.ColorWhite,
	"input_modal_field_bg":   tcell.ColorBlue,
	"input_modal_field_text": tcell.ColorWhite,

	"bkmk_modal_bg":         tcell.ColorTeal,
	"bkmk_modal_text":       tcell.ColorWhite,
	"bkmk_modal_label":      tcell.ColorYellow,
	"bkmk_modal_field_bg":   tcell.ColorBlue,
	"bkmk_modal_field_text": tcell.ColorWhite,

	"hdg_1":             tcell.ColorRed,
	"hdg_2":             tcell.ColorLime,
	"hdg_3":             tcell.ColorFuchsia,
	"amfora_link":       tcell.Color33, // xterm:DodgerBlue1, #0087ff
	"foreign_link":      tcell.Color92, // xterm:DarkViolet, #8700d7
	"link_number":       tcell.ColorSilver,
	"regular_text":      tcell.ColorWhite,
	"quote_text":        tcell.ColorWhite,
	"preformatted_text": tcell.Color229, // xterm:Wheat1, #ffffaf
	"list_text":         tcell.ColorWhite,
}

func SetColor(key string, color tcell.Color) {
	themeMu.Lock()
	theme[key] = color
	themeMu.Unlock()
}

// GetColor will return tcell.ColorBlack if there is no color for the provided key.
func GetColor(key string) tcell.Color {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return theme[key].TrueColor()
}

// GetColorString returns a string that can be used in a cview color tag,
// for the given theme key.
// It will return "#000000" if there is no color for the provided key.
func GetColorString(key string) string {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return fmt.Sprintf("#%06x", theme[key].TrueColor().Hex())
}
