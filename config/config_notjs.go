//go:build !js
// +build !js

package config

import (
	"github.com/muesli/termenv"
)

func init() {
	hasDarkTerminalBackground = termenv.HasDarkBackground()
}
