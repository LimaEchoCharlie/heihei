package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var loc [2]float64 // the [latitude, longitude] of the device

const configFilename = "configuration.json"

// loadConfiguration loads the server configuration
func loadConfiguration() error {
	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	path := filepath.Dir(ex)

	file, err := os.Open(filepath.Join(path, configFilename))
	if err != nil {
		return err
	}

	// use pointers for required values
	config := struct {
		Location *[2]float64
	}{}
	// decode json
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	if config.Location == nil {
		logger.Panicf("Location is missing from %v\n", configFilename)
	}
	loc = *config.Location
	return nil
}

// location returns the latitude and longitude of the device
func location() (float64, float64) {
	return loc[0], loc[1]
}
