// +build linux freebsd netbsd openbsd

package webbrowser

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenInBrowser checks for the presence of a display server
// and environment variables indicating a gui is present. If found
// then xdg-open is called on a url to open said url in the default
// gui web browser for the system
func Open(url string) (string, error) {
	disp := os.Getenv("DISPLAY")
	wayland := os.Getenv("WAYLAND_DISPLAY")
	_, err := exec.LookPath("Xorg")
	if disp == "" && wayland == "" && err != nil {
		return "", fmt.Errorf("no gui is available")
	}

	_, err = exec.LookPath("xdg-open")
	if err != nil {
		return "", fmt.Errorf("xdg-open command not found, cannot open in web browser")
	}
	// Use start rather than run or output in order
	// to release the process and not block
	err = exec.Command("xdg-open", url).Start()
	if err != nil {
		return "", err
	}
	return "Opened in system default web browser", nil
}
