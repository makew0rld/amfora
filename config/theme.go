package config

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Functions to allow themeing configuration.
// UI element tcell.Colors are mapped to a string key, such as "error" or "tab_bg"
// These are the same keys used in the config file.

// Special color with no real color value
// Used for a default foreground color
// White is the terminal background is black, black if the terminal background is white
// Converted to a real color in this file before being sent out to other modules
const ColorFg = tcell.ColorSpecial | 2

// The same as ColorFg, but inverted
const ColorBg = tcell.ColorSpecial | 3

var themeMu = sync.RWMutex{}
var theme = map[string]tcell.Color{
	// Map these for special uses in code
	"ColorBg": ColorBg,
	"ColorFg": ColorFg,

	// Default values below
	// Only the 16 Xterm system tcell.Colors are used, because those are the tcell.Colors overridden
	// by the user's default terminal theme

	// Used for cview.Styles.PrimitiveBackgroundColor
	// Set to tcell.ColorDefault because that allows transparent terminals to work
	// The rest of this theme assumes that the background is equivalent to black, but
	// white colors switched to black later if the background is determined to be white.
	//
	// Also, this is set to tcell.ColorBlack in config.go if colors are disabled in the config.
	"bg": tcell.ColorDefault,

	"tab_num":         tcell.ColorTeal,
	"tab_divider":     ColorFg,
	"bottombar_label": tcell.ColorTeal,
	"bottombar_text":  ColorBg,
	"bottombar_bg":    ColorFg,
	"scrollbar":       ColorFg,

	// Modals
	"btn_bg":   tcell.ColorTeal,  // All modal buttons
	"btn_text": tcell.ColorWhite, // White instead of ColorFg because background is known to be Teal

	"dl_choice_modal_bg":      tcell.ColorOlive,
	"dl_choice_modal_text":    tcell.ColorWhite,
	"dl_modal_bg":             tcell.ColorOlive,
	"dl_modal_text":           tcell.ColorWhite,
	"info_modal_bg":           tcell.ColorGray,
	"info_modal_text":         tcell.ColorWhite,
	"error_modal_bg":          tcell.ColorMaroon,
	"error_modal_text":        tcell.ColorWhite,
	"yesno_modal_bg":          tcell.ColorTeal,
	"yesno_modal_text":        tcell.ColorWhite,
	"tofu_modal_bg":           tcell.ColorMaroon,
	"tofu_modal_text":         tcell.ColorWhite,
	"subscription_modal_bg":   tcell.ColorTeal,
	"subscription_modal_text": tcell.ColorWhite,

	"input_modal_bg":         tcell.ColorGreen,
	"input_modal_text":       tcell.ColorWhite,
	"input_modal_field_bg":   tcell.ColorNavy,
	"input_modal_field_text": tcell.ColorWhite,

	"bkmk_modal_bg":         tcell.ColorTeal,
	"bkmk_modal_text":       tcell.ColorWhite,
	"bkmk_modal_label":      tcell.ColorYellow,
	"bkmk_modal_field_bg":   tcell.ColorNavy,
	"bkmk_modal_field_text": tcell.ColorWhite,

	"hdg_1":             tcell.ColorRed,
	"hdg_2":             tcell.ColorLime,
	"hdg_3":             tcell.ColorFuchsia,
	"amfora_link":       tcell.ColorBlue,
	"foreign_link":      tcell.ColorPurple,
	"link_number":       tcell.ColorSilver,
	"regular_text":      ColorFg,
	"quote_text":        ColorFg,
	"preformatted_text": ColorFg,
	"list_text":         ColorFg,
}

func SetColor(key string, color tcell.Color) {
	themeMu.Lock()
	// Use truecolor because this is only called with user-set tcell.Colors
	// Which should be represented exactly
	theme[key] = color.TrueColor()
	themeMu.Unlock()
}

// GetColor will return tcell.ColorBlack if there is no tcell.Color for the provided key.
func GetColor(key string) tcell.Color {
	themeMu.RLock()
	defer themeMu.RUnlock()

	color := theme[key]

	if color == ColorFg {
		if hasDarkTerminalBackground {
			return tcell.ColorWhite
		}
		return tcell.ColorBlack
	}
	if color == ColorBg {
		if hasDarkTerminalBackground {
			return tcell.ColorBlack
		}
		return tcell.ColorWhite
	}

	return color
}

// colorToString converts a color to a string for use in a cview tag
func colorToString(color tcell.Color) string {
	if color == tcell.ColorDefault {
		return "-"
	}

	if color == ColorFg {
		if hasDarkTerminalBackground {
			return "white"
		}
		return "black"
	}
	if color == ColorBg {
		if hasDarkTerminalBackground {
			return "black"
		}
		return "white"
	}

	if color&tcell.ColorIsRGB == 0 {
		// tcell.Color is not RGB/TrueColor, it's a tcell.Color from the default terminal
		// theme as set above
		// Return a tcell.Color name instead of a hex code, so that cview doesn't use TrueColor
		return ColorToColorName[color]
	}

	// Color set by user, must be respected exactly so hex code is used
	return fmt.Sprintf("#%06x", color.Hex())
}

// GetColorString returns a string that can be used in a cview tcell.Color tag,
// for the given theme key.
// It will return "#000000" if there is no tcell.Color for the provided key.
func GetColorString(key string) string {
	themeMu.RLock()
	defer themeMu.RUnlock()

	return colorToString(theme[key])
}

// GetContrastingColor returns tcell.ColorBlack if tcell.Color is brighter than gray
// otherwise returns tcell.ColorWhite if tcell.Color is dimmer than gray
// if tcell.Color is tcell.ColorDefault (undefined luminance) this returns tcell.ColorDefault
func GetContrastingColor(color tcell.Color) tcell.Color {
	if color == tcell.ColorDefault {
		// tcell.Color should never be tcell.ColorDefault
		// only config keys which end in bg are allowed to be set to default
		// and the only way the argument of this function is set to tcell.ColorDefault
		// is if both the text and bg of an element in the UI are set to default
		return tcell.ColorDefault
	}
	r, g, b := color.RGB()
	luminance := (77*r + 150*g + 29*b + 1<<7) >> 8
	const gray = 119 // The middle gray
	if luminance > gray {
		return tcell.ColorBlack
	}
	return tcell.ColorWhite
}

// GetTextColor is the Same as GetColor, unless the key is "default".
// This happens on focus of a UI element which has a bg of default, in which case
// It return tcell.ColorBlack or tcell.ColorWhite, depending on which is more readable
func GetTextColor(key, bg string) tcell.Color {
	themeMu.RLock()
	defer themeMu.RUnlock()
	color := theme[key].TrueColor()
	if color != tcell.ColorDefault {
		return color
	}
	return GetContrastingColor(theme[bg].TrueColor())
}

// GetTextColorString is the Same as GetColorString, unless the key is "default".
// This happens on focus of a UI element which has a bg of default, in which case
// It return tcell.ColorBlack or tcell.ColorWhite, depending on which is more readable
func GetTextColorString(key, bg string) string {
	return colorToString(GetTextColor(key, bg))
}

// Inverted version of a tcell map
// https://github.com/gdamore/tcell/blob/v2.3.3/color.go#L845
var ColorToColorName = map[tcell.Color]string{
	tcell.ColorBlack:                "black",
	tcell.ColorMaroon:               "maroon",
	tcell.ColorGreen:                "green",
	tcell.ColorOlive:                "olive",
	tcell.ColorNavy:                 "navy",
	tcell.ColorPurple:               "purple",
	tcell.ColorTeal:                 "teal",
	tcell.ColorSilver:               "silver",
	tcell.ColorGray:                 "gray",
	tcell.ColorRed:                  "red",
	tcell.ColorLime:                 "lime",
	tcell.ColorYellow:               "yellow",
	tcell.ColorBlue:                 "blue",
	tcell.ColorFuchsia:              "fuchsia",
	tcell.ColorAqua:                 "aqua",
	tcell.ColorWhite:                "white",
	tcell.ColorAliceBlue:            "aliceblue",
	tcell.ColorAntiqueWhite:         "antiquewhite",
	tcell.ColorAquaMarine:           "aquamarine",
	tcell.ColorAzure:                "azure",
	tcell.ColorBeige:                "beige",
	tcell.ColorBisque:               "bisque",
	tcell.ColorBlanchedAlmond:       "blanchedalmond",
	tcell.ColorBlueViolet:           "blueviolet",
	tcell.ColorBrown:                "brown",
	tcell.ColorBurlyWood:            "burlywood",
	tcell.ColorCadetBlue:            "cadetblue",
	tcell.ColorChartreuse:           "chartreuse",
	tcell.ColorChocolate:            "chocolate",
	tcell.ColorCoral:                "coral",
	tcell.ColorCornflowerBlue:       "cornflowerblue",
	tcell.ColorCornsilk:             "cornsilk",
	tcell.ColorCrimson:              "crimson",
	tcell.ColorDarkBlue:             "darkblue",
	tcell.ColorDarkCyan:             "darkcyan",
	tcell.ColorDarkGoldenrod:        "darkgoldenrod",
	tcell.ColorDarkGray:             "darkgray",
	tcell.ColorDarkGreen:            "darkgreen",
	tcell.ColorDarkKhaki:            "darkkhaki",
	tcell.ColorDarkMagenta:          "darkmagenta",
	tcell.ColorDarkOliveGreen:       "darkolivegreen",
	tcell.ColorDarkOrange:           "darkorange",
	tcell.ColorDarkOrchid:           "darkorchid",
	tcell.ColorDarkRed:              "darkred",
	tcell.ColorDarkSalmon:           "darksalmon",
	tcell.ColorDarkSeaGreen:         "darkseagreen",
	tcell.ColorDarkSlateBlue:        "darkslateblue",
	tcell.ColorDarkSlateGray:        "darkslategray",
	tcell.ColorDarkTurquoise:        "darkturquoise",
	tcell.ColorDarkViolet:           "darkviolet",
	tcell.ColorDeepPink:             "deeppink",
	tcell.ColorDeepSkyBlue:          "deepskyblue",
	tcell.ColorDimGray:              "dimgray",
	tcell.ColorDodgerBlue:           "dodgerblue",
	tcell.ColorFireBrick:            "firebrick",
	tcell.ColorFloralWhite:          "floralwhite",
	tcell.ColorForestGreen:          "forestgreen",
	tcell.ColorGainsboro:            "gainsboro",
	tcell.ColorGhostWhite:           "ghostwhite",
	tcell.ColorGold:                 "gold",
	tcell.ColorGoldenrod:            "goldenrod",
	tcell.ColorGreenYellow:          "greenyellow",
	tcell.ColorHoneydew:             "honeydew",
	tcell.ColorHotPink:              "hotpink",
	tcell.ColorIndianRed:            "indianred",
	tcell.ColorIndigo:               "indigo",
	tcell.ColorIvory:                "ivory",
	tcell.ColorKhaki:                "khaki",
	tcell.ColorLavender:             "lavender",
	tcell.ColorLavenderBlush:        "lavenderblush",
	tcell.ColorLawnGreen:            "lawngreen",
	tcell.ColorLemonChiffon:         "lemonchiffon",
	tcell.ColorLightBlue:            "lightblue",
	tcell.ColorLightCoral:           "lightcoral",
	tcell.ColorLightCyan:            "lightcyan",
	tcell.ColorLightGoldenrodYellow: "lightgoldenrodyellow",
	tcell.ColorLightGray:            "lightgray",
	tcell.ColorLightGreen:           "lightgreen",
	tcell.ColorLightPink:            "lightpink",
	tcell.ColorLightSalmon:          "lightsalmon",
	tcell.ColorLightSeaGreen:        "lightseagreen",
	tcell.ColorLightSkyBlue:         "lightskyblue",
	tcell.ColorLightSlateGray:       "lightslategray",
	tcell.ColorLightSteelBlue:       "lightsteelblue",
	tcell.ColorLightYellow:          "lightyellow",
	tcell.ColorLimeGreen:            "limegreen",
	tcell.ColorLinen:                "linen",
	tcell.ColorMediumAquamarine:     "mediumaquamarine",
	tcell.ColorMediumBlue:           "mediumblue",
	tcell.ColorMediumOrchid:         "mediumorchid",
	tcell.ColorMediumPurple:         "mediumpurple",
	tcell.ColorMediumSeaGreen:       "mediumseagreen",
	tcell.ColorMediumSlateBlue:      "mediumslateblue",
	tcell.ColorMediumSpringGreen:    "mediumspringgreen",
	tcell.ColorMediumTurquoise:      "mediumturquoise",
	tcell.ColorMediumVioletRed:      "mediumvioletred",
	tcell.ColorMidnightBlue:         "midnightblue",
	tcell.ColorMintCream:            "mintcream",
	tcell.ColorMistyRose:            "mistyrose",
	tcell.ColorMoccasin:             "moccasin",
	tcell.ColorNavajoWhite:          "navajowhite",
	tcell.ColorOldLace:              "oldlace",
	tcell.ColorOliveDrab:            "olivedrab",
	tcell.ColorOrange:               "orange",
	tcell.ColorOrangeRed:            "orangered",
	tcell.ColorOrchid:               "orchid",
	tcell.ColorPaleGoldenrod:        "palegoldenrod",
	tcell.ColorPaleGreen:            "palegreen",
	tcell.ColorPaleTurquoise:        "paleturquoise",
	tcell.ColorPaleVioletRed:        "palevioletred",
	tcell.ColorPapayaWhip:           "papayawhip",
	tcell.ColorPeachPuff:            "peachpuff",
	tcell.ColorPeru:                 "peru",
	tcell.ColorPink:                 "pink",
	tcell.ColorPlum:                 "plum",
	tcell.ColorPowderBlue:           "powderblue",
	tcell.ColorRebeccaPurple:        "rebeccapurple",
	tcell.ColorRosyBrown:            "rosybrown",
	tcell.ColorRoyalBlue:            "royalblue",
	tcell.ColorSaddleBrown:          "saddlebrown",
	tcell.ColorSalmon:               "salmon",
	tcell.ColorSandyBrown:           "sandybrown",
	tcell.ColorSeaGreen:             "seagreen",
	tcell.ColorSeashell:             "seashell",
	tcell.ColorSienna:               "sienna",
	tcell.ColorSkyblue:              "skyblue",
	tcell.ColorSlateBlue:            "slateblue",
	tcell.ColorSlateGray:            "slategray",
	tcell.ColorSnow:                 "snow",
	tcell.ColorSpringGreen:          "springgreen",
	tcell.ColorSteelBlue:            "steelblue",
	tcell.ColorTan:                  "tan",
	tcell.ColorThistle:              "thistle",
	tcell.ColorTomato:               "tomato",
	tcell.ColorTurquoise:            "turquoise",
	tcell.ColorViolet:               "violet",
	tcell.ColorWheat:                "wheat",
	tcell.ColorWhiteSmoke:           "whitesmoke",
	tcell.ColorYellowGreen:          "yellowgreen",
}
