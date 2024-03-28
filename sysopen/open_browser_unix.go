//go:build linux || freebsd || netbsd || openbsd
// +build linux freebsd netbsd openbsd

//nolint:goerr113
package sysopen

import (
	"fmt"
	"os"
	"os/exec"
)

// Open opens `path` in default system viewer. It tries to do so using
// xdg-open. It only works if there is a display server working.
func Open(path string) (string, error) {
	var (
		xorgDisplay                     = os.Getenv("DISPLAY")
		waylandDisplay                  = os.Getenv("WAYLAND_DISPLAY")
		xdgOpenPath, xdgOpenNotFoundErr = exec.LookPath("xdg-open")
	)
	switch {
	case xorgDisplay == "" && waylandDisplay == "":
		return "", fmt.Errorf("no display server was found. " +
			"You may set a default command in the config")
	case xdgOpenNotFoundErr == nil:
		// Use start rather than run or output in order
		// to make application run in background.
		proc := exec.Command(xdgOpenPath, path)
		if err := proc.Start(); err != nil {
			return "", err
		}
		//nolint:errcheck
		go proc.Wait() // Prevent zombies, see #219
		return "Opened in default system viewer", nil
	default:
		return "", fmt.Errorf("could not determine default system viewer. " +
			"Set a catch-all command in the config")
	}
}
