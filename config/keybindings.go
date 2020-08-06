package config

import (
	"errors"

	"github.com/spf13/viper"
)

// KeyToNum returns the number on the user's keyboard they pressed,
// using the rune returned when when they press Shift+Num.
// The error is not nil if the provided key is invalid.
func KeyToNum(key rune) (int, error) {
	runes := []rune(viper.GetString("keybindings.shift_numbers"))
	for i := range runes {
		if key == runes[i] {
			if i == len(runes)-1 {
				// Last key is 0, not 10
				return 0, nil
			}
			return i + 1, nil
		}
	}
	return -1, errors.New("provided key is invalid")
}
