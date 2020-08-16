package main

import (
	"fmt"
	"os"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
	"github.com/makeworld-the-better-one/amfora/feeds"
)

var version = "1.5.0-unreleased"

func main() {
	// err := logger.Init()
	// if err != nil {
	// 	panic(err)
	// }

	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Println("amfora v" + version)
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
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	err = feeds.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "feeds.json error: %v\n", err)
		os.Exit(1)
	}

	display.Init()
	display.NewTab()
	display.NewTab() // Open extra tab and close it to fully initialize the app and wrapping
	display.CloseTab()
	if len(os.Args[1:]) > 0 {
		display.URL(os.Args[1])
	}

	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
