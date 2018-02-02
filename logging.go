package main

import (
	"log"
	"os"
	"path/filepath"
)

var (
	logger *log.Logger
	file   *os.File
)

// initLogging creates a new logger that logs to a file in the executable directory
func initLogging() error {
	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	path := filepath.Dir(ex)

	// If the logfile doesn't exist, create it. Otherwise append to it.
	file, err := os.OpenFile(filepath.Join(path, "heihei.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

// stopLogging stops logging and clears up
func stopLogging() {
	logger = nil
	file.Close()
}
