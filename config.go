package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var config = configuration{}

type configuration struct {
	Location [2]float64 // the [latitude, longitude] of the device
}

// loadConfiguration loads the server configuration
func loadConfiguration() error {
	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	path := filepath.Dir(ex)

	file, err := os.Open(filepath.Join(path, "configuration.json"))
	if err != nil {
		return err
	}

	// decode json
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

// location returns the latitude and longitude of the device
func location() (float64, float64) {
	return config.Location[0], config.Location[1]
}
