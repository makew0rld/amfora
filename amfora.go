package main

import (
	"fmt"
	"os"

	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
	"github.com/makeworld-the-better-one/amfora/subscriptions"
)

var (
	version = "v1.7.2"
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
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	err = subscriptions.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "subscriptions.json error: %v\n", err)
		os.Exit(1)
	}

	client.Init()

	display.Init(version, commit, builtBy)
	display.NewTab()
	display.NewTab() // Open extra tab and close it to fully initialize the app and wrapping
	display.CloseTab()
	if len(os.Args[1:]) > 0 {
		display.SearchOrLoad(os.Args[1], -1)
	}

	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
