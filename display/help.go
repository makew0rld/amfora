package display

import "strings"

var helpCells = strings.TrimSpace(`
?|Bring up this help.
Esc|Leave the help
Arrow keys, h/j/k/l|Scroll and move a page.
Tab|Navigate to the next item in a popup.
Shift-Tab|Navigate to the previous item in a popup.
Ctrl-H|Go home
Ctrl-T|New tab
Ctrl-W|Close tab. For now, only the right-most tab can be closed.
b|Go back a page
f|Go forward a page
g|Go to top of document
G|Go to bottom of document
spacebar|Open bar at the bottom - type a URL or link number
Enter|On a page this will start link highlighting. Press Tab and Shift-Tab to pick different links. Press enter again to go to one.
Ctrl-R, R|Reload a page. This also clears the cache.
q, Ctrl-Q|Quit
Shift-NUMBER|Go to a specific tab.
Shift-0, )|Go to the last tab.
`)
