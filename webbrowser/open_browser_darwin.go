// +build darwin

package webbrowser

import "os/exec"

// Open opens `url` in default system browser.
func Open(url string) (string, error) {
	err := exec.Command("open", url).Start()
	if err != nil {
		return "", err
	}
	return "Opened in system default web browser", nil
}
