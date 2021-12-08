package logger

// For debugging

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var logger *log.Logger

func GetLogger() (*log.Logger, error) {
	if logger != nil {
		return logger, nil
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

	logger = log.New(writer, "", log.LstdFlags)

	if !debugModeEnabled {
		// Clear all flags to skip log output formatting step to increase
		// performance somewhat if we're not logging anything
		logger.SetFlags(0)
	}

	logger.Println("Started logger")

	return logger, nil
}
