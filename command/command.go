package command

import (
	"os/exec"
	"strings"
)

// RunCommand runs `command`, replacing the string "${url}" with `url`.
func RunCommand(command string, url string) (string, error) {
	cmdWithURL := strings.ReplaceAll(command, "${url}", url)
	cmdSplit := strings.SplitN(cmdWithURL, " ", 2)
	if len(cmdSplit) > 1 {
		if err := exec.Command(cmdSplit[0], cmdSplit[1]).Start(); err != nil {
			return "", err
		}
		return "Ran command " + cmdSplit[0] + " with args " + cmdSplit[1], nil
	}
	if err := exec.Command(cmdWithURL).Start(); err != nil {
		return "", err
	}
	return "Ran command " + cmdWithURL, nil
}
