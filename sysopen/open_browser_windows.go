//go:build windows && (!linux || !darwin || !freebsd || !netbsd || !openbsd)
// +build windows
// +build !linux !darwin !freebsd !netbsd !openbsd

package sysopen

import "os/exec"

// Open opens `path` in default system vierwer.
func Open(path string) (string, error) {
	proc := exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	err := proc.Start()
	if err != nil {
		return "", err
	}
	go proc.Wait() // Prevent zombies, see #219
	return "Opened in default system viewer", nil
}
