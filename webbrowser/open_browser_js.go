// +build js

package webbrowser

import "errors"
import "syscall/js"

// Open opens `url` in default system browser, but not on this OS.
func Open(url string) (string, error) {
	if newWindow := js.Global().Call("open", url); !newWindow.IsNull() {
		return "Opened in new tab", nil
	} else {
		return "", errors.New("Browser refused to open new tab")
	}
}

