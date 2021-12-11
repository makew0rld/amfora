package display

import (
	"strconv"

	"github.com/makeworld-the-better-one/amfora/command"
	"github.com/spf13/viper"
)

// CustomCommand runs custom commands as defined in the app configuration.
// Commands are zero-indexed, so 0 is command1 and 9 is command0 (10).
func CustomCommand(num int, url string) {
	if num < 0 {
		num = 0
	}
	num++
	if num > 9 {
		num = 0
	}

	cmd := viper.GetString("commands.command" + strconv.Itoa(num))
	if len(cmd) > 0 {
		msg, err := command.RunCommand(cmd, url)
		if err != nil {
			Error("Command Error", err.Error())
			return
		}
		Info(msg)
	} else {
		Error("Command Error", "Command "+strconv.Itoa(num)+" not defined")
		return
	}

	App.Draw()
}
