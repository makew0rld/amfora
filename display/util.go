package display

import (
	"errors"
	"net/url"
	"strings"

	"code.rocketnine.space/tslocum/cview"
	"github.com/spf13/viper"
)

// This file contains funcs that are small, self-contained utilities.

// makeContentLayout returns a flex that contains the given TextView
// along with the provided left margin, as well as a single empty
// line at the top, for a top margin.
func makeContentLayout(tv *cview.TextView, leftMargin int) *cview.Flex {
	// Create horizontal flex with the left margin as an empty space
	horiz := cview.NewFlex()
	horiz.SetDirection(cview.FlexColumn)
	if leftMargin > 0 {
		horiz.AddItem(nil, leftMargin, 0, false)
	}
	horiz.AddItem(tv, 0, 1, true)

	// Create a vertical flex with the other one and a top margin
	vert := cview.NewFlex()
	vert.SetDirection(cview.FlexRow)
	vert.AddItem(nil, 1, 0, false)
	vert.AddItem(horiz, 0, 1, true)

	return vert
}

// makeTabLabel takes a string and adds spacing to it, making it
// suitable for display as a tab label.
func makeTabLabel(s string) string {
	return " " + s + " "
}

// tabNumber gets the index of the tab in the tabs slice. It returns -1
// if the tab is not in that slice.
func tabNumber(t *tab) int {
	tempTabs := tabs
	for i := range tempTabs {
		if tempTabs[i] == t {
			return i
		}
	}
	return -1
}

// escapeMeta santizes a META string for use within a cview modal.
func escapeMeta(meta string) string {
	return cview.Escape(strings.ReplaceAll(meta, "\n", ""))
}

// isValidTab indicates whether the passed tab is still being used, even if it's not currently displayed.
func isValidTab(t *tab) bool {
	return tabNumber(t) != -1
}

func leftMargin() int {
	// Return the left margin size that centers the text, assuming it's the max width
	// https://github.com/makeworld-the-better-one/amfora/issues/233

	lm := (termW - viper.GetInt("a-general.max_width")) / 2
	if lm < 0 {
		return 0
	}
	return lm
}

func textWidth() int {
	if termW <= 0 {
		// This prevent a flash of 1-column text on startup, when the terminal
		// width hasn't been initialized.
		return viper.GetInt("a-general.max_width")
	}

	// Subtract left and right margin from total width to get text width
	// Left and right margin are equal because text is automatically centered, see:
	// https://github.com/makeworld-the-better-one/amfora/issues/233

	max := termW - leftMargin()*2
	if max < viper.GetInt("a-general.max_width") {
		return max
	}
	return viper.GetInt("a-general.max_width")
}

// resolveRelLink returns an absolute link for the given absolute link and relative one.
// It also returns an error if it could not resolve the links, which should be displayed
// to the user.
func resolveRelLink(t *tab, prev, next string) (string, error) {
	if !t.hasContent() || t.isAnAboutPage() {
		return next, nil
	}

	prevParsed, _ := url.Parse(prev)
	nextParsed, err := url.Parse(next)
	if err != nil {
		return "", errors.New("link URL could not be parsed") //nolint:goerr113
	}
	return prevParsed.ResolveReference(nextParsed).String(), nil
}
