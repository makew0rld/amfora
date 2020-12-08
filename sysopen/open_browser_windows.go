// +build windows
// +build !linux !darwin !freebsd !netbsd !openbsd

package sysopen

import "os/exec"

// Open opens `path` in default system vierwer.
func Open(path string) (string, error) {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	if err != nil {
		return "", err
	}
	return "Opened in default system viewer", nil
}
