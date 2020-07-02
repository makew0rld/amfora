// +build !linux,!darwin,!windows,!freebsd,!netbsd,!openbsd

package webbrowser

import "fmt"

func Open(url string) (string, error) {
	return "", fmt.Errorf("unsupported OS for default HTTP handling. Set a command in the config")
}
