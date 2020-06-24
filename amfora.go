package main

import (
	"fmt"
	"os"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
)

var version = "1.2.0"

func main() {
	// err := logger.Init()
	// if err != nil {
	// 	panic(err)
	// }

	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Println(version)
			return
		}
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Println("Amfora is a fancy terminal browser for the Gemini protocol.\r\n")
			fmt.Println("Usage:\r\namfora [URL]\r\namfora --version, -v")
			return
		}
	}

	err := config.Init()
	if err != nil {
		panic(err)
	}
	display.Init()

	if len(os.Args[1:]) == 0 {
		// There should always be a tab
		display.NewTab()
	} else {
		display.NewTab()
		display.URL(os.Args[1])
	}

	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
