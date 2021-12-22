//go:build darwin
// +build darwin

package sysopen

import "os/exec"

// Open opens `path` in default system viewer.
func Open(path string) (string, error) {
	proc := exec.Command("open", path)
	err := proc.Start()
	if err != nil {
		return "", err
	}
	go proc.Wait() // Prevent zombies, see #219
	return "Opened in default system viewer", nil
}
