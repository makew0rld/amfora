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
