package logger

// For debugging

import (
	"log"
	"os"
)

var Log *log.Logger

func Init() error {
	f, err := os.Create("debug.log")
	if err != nil {
		return err
	}
	Log = log.New(f, "", log.LstdFlags)
	Log.Println("Started Log")
	return nil
}
