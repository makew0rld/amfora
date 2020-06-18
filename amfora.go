package main

import (
	"os"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/display"
)

func main() {
	// err := logger.Init()
	// if err != nil {
	// 	panic(err)
	// }

	err := config.Init()
	if err != nil {
		panic(err)
	}
	display.Init()

	for _, url := range os.Args[1:] {
		display.NewTab()
		display.URL(url)
	}

	if len(os.Args[1:]) == 0 {
		// There should always be a tab
		display.NewTab()
	}

	if err = display.App.Run(); err != nil {
		panic(err)
	}
}
