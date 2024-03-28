package logger

// For debugging

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var Logger *log.Logger

func GetLogger() (*log.Logger, error) {
	if Logger != nil {
		return Logger, nil
	}

	var writer io.Writer
	var err error

	debugModeEnabled := os.Getenv("AMFORA_DEBUG") == "1"
	if debugModeEnabled {
		writer, err = os.Create("debug.log")
		if err != nil {
			return nil, err
		}
	} else {
		// Suppress all logging output if debug mode is disabled
		writer = ioutil.Discard
	}

	Logger = log.New(writer, "", log.LstdFlags)

	if !debugModeEnabled {
		// Clear all flags to skip log output formatting step to increase
		// performance somewhat if we're not logging anything
		Logger.SetFlags(0)
	}

	Logger.Println("Started logger")

	return Logger, nil
}
