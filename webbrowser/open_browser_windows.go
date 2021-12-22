//go:build windows && (!linux || !darwin || !freebsd || !netbsd || !openbsd)
// +build windows
// +build !linux !darwin !freebsd !netbsd !openbsd

package webbrowser

import "os/exec"

// Open opens `url` in default system browser.
func Open(url string) (string, error) {
	proc := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	err := proc.Start()
	if err != nil {
		return "", err
	}
	go proc.Wait() // Prevent zombies, see #219
	return "Opened in system default web browser", nil
}
