//go:build linux || freebsd || netbsd || openbsd
// +build linux freebsd netbsd openbsd

//nolint:goerr113
package webbrowser

import (
	"fmt"
	"os"
	"os/exec"
)

// Open opens `url` in default system browser. It tries to do so in two
// ways (xdg-open and $BROWSER). It only works if there is a display
// server working.
//
// bouncepaw: I tried to support TTYs as well. The idea was to open
// a browser in foreground and return back to amfora after the browser
// is closed. While all browsers I tested opened correctly (w3m, lynx),
// I couldn't make it restore amfora correctly. The screen always ended
// up distorted. None of my stunts with altscreen buffers helped.
func Open(url string) (string, error) {
	var (
		// In prev versions there was also Xorg executable checked for.
		// I don't see any reason to check for it.
		xorgDisplay                     = os.Getenv("DISPLAY")
		waylandDisplay                  = os.Getenv("WAYLAND_DISPLAY")
		xdgOpenPath, xdgOpenNotFoundErr = exec.LookPath("xdg-open")
		envBrowser                      = os.Getenv("BROWSER")
	)
	switch {
	case xorgDisplay == "" && waylandDisplay == "":
		return "", fmt.Errorf("no display server was found")
	case xdgOpenNotFoundErr == nil: // Prefer xdg-open over $BROWSER
		// Use start rather than run or output in order
		// to make browser running in background.
		proc := exec.Command(xdgOpenPath, url)
		if err := proc.Start(); err != nil {
			return "", err
		}
		go proc.Wait() // Prevent zombies, see #219
		return "Opened in system default web browser", nil
	case envBrowser != "":
		proc := exec.Command(envBrowser, url)
		if err := proc.Start(); err != nil {
			return "", err
		}
		go proc.Wait() // Prevent zombies, see #219
		return "Opened in system default web browser", nil
	default:
		return "", fmt.Errorf("could not determine system browser")
	}
}
