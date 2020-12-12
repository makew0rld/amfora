// +build !linux,!darwin,!windows,!freebsd,!netbsd,!openbsd

package sysopen

import "fmt"

// Open opens `path` in default system viewer, but not on this OS.
func Open(path string) (string, error) {
	return "", fmt.Errorf("unsupported OS for default system viewer. " +
		"Set a catch-all [[mediatype-handlers]] command in the config")
}
