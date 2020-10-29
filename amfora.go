package main

import (
	"fmt"
	"os"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
)

var (
	version = "1.5.0"
	commit  = "unknown"
	builtBy = "unknown"
)

func main() {
	// err := logger.Init()
	// if err != nil {
	// 	panic(err)
	// }

	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Println("Amfora", version)
			fmt.Println("Commit:", commit)
			fmt.Println("Built by:", builtBy)
			return
		}
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Println("Amfora is a fancy terminal browser for the Gemini protocol.")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("amfora [URL]")
			fmt.Println("amfora --version, -v")
			return
		}
	}

	err := config.Init()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}

	// Initalize lower-level cview app
	if err = display.App.Init(); err != nil {
		panic(err)
	}

	// Initialize Amfora's settings
	display.Init()
	display.NewTab()
	if len(os.Args[1:]) > 0 {
		display.URL(os.Args[1])
	}

	// Start
	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
