package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/makeworld-the-better-one/amfora/bookmarks"
	"github.com/makeworld-the-better-one/amfora/client"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/amfora/subscriptions"
)

var (
	version = "v1.10.0"
	commit  = "unknown"
	builtBy = "unknown"
)

func main() {
	log, err := logger.GetLogger()
	if err != nil {
		panic(err)
	}

	debugModeEnabled := os.Getenv("AMFORA_DEBUG") == "1"
	if debugModeEnabled {
		log.Println("Debug mode enabled")
	}

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

	err = config.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	err = client.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Client error: %v\n", err)
		os.Exit(1)
	}

	err = subscriptions.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "subscriptions.json error: %v\n", err)
		os.Exit(1)
	}
	err = bookmarks.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bookmarks.xml error: %v\n", err)
		os.Exit(1)
	}

	// Initialize lower-level cview app
	if err = display.App.Init(); err != nil {
		panic(err)
	}

	// Initialize Amfora's settings
	display.Init(version, commit, builtBy)

	// Load a URL, file, or render from stdin
	if len(os.Args[1:]) > 0 {
		url := os.Args[1]
		if !strings.Contains(url, "://") || strings.HasPrefix(url, "../") || strings.HasPrefix(url, "./") {
			fileName := url
			if _, err := os.Stat(fileName); err == nil {
				if !strings.HasPrefix(fileName, "/") {
					cwd, err := os.Getwd()
					if err != nil {
						fmt.Fprintf(os.Stderr, "error getting working directory path: %v\n", err)
						os.Exit(1)
					}
					fileName = filepath.Join(cwd, fileName)
				}
				url = "file://" + fileName
			}
		}
		display.NewTabWithURL(url)
	} else if !isStdinEmpty() {
		display.NewTab()
		renderFromStdin()
	} else {
		display.NewTab()
	}

	// Start
	if err = display.App.Run(); err != nil {
		panic(err)
	}
}

func isStdinEmpty() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func renderFromStdin() {
	stdinTextBuilder := new(strings.Builder)
	_, err := io.Copy(stdinTextBuilder, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading from standard input: %v\n", err)
		os.Exit(1)
	}

	stdinText := stdinTextBuilder.String()
	display.RenderFromString(stdinText)
}
