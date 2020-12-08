// +build darwin

package sysopen

import "os/exec"

// Open opens `path` in default system viewer.
func Open(path string) (string, error) {
	err := exec.Command("open", path).Start()
	if err != nil {
		return "", err
	}
	return "Opened in default system viewer", nil
}
