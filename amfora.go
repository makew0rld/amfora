package main

import (
	"fmt"
	"os"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
)

var version = "1.1.0"

func main() {
	// err := logger.Init()
	// if err != nil {
	// 	panic(err)
	// }

	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Print(version + "\r\n")
			return
		}
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Print("Amfora is a fancy terminal browser for the Gemini protocol.\r\n\r\n")
			fmt.Print("Usage:\r\namfora [URL]\r\namfora --version, -v\r\n")
			return
		}
	}

	err := config.Init()
	if err != nil {
		panic(err)
	}
	display.Init()

	display.NewTab()
	if len(os.Args[1:]) > 0 {
		display.URL(os.Args[1])
	}

	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
