package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var loc [2]float64 // the [latitude, longitude] of the device
var lightsOut string

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
		Location  *[2]float64 `json:"location"`
		LightsOut *string     `json:"lights_out"`
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

	// check that a lights out value has been supplied and that it has a valid syntax
	if config.LightsOut == nil {
		logger.Panicf("lights out is missing from %v\n", configFilename)
	} else if _, _, err = decodeClock(*config.LightsOut); err != nil {
		logger.Panicf("lights out value from %v: %v\n", configFilename, err)
	}
	lightsOut = *config.LightsOut

	return nil
}

// location returns the latitude and longitude of the device
func location() (float64, float64) {
	return loc[0], loc[1]
}

var pattern = regexp.MustCompile("^([0-9]{1,2}):([0-9]{2})$")

// decodeClock converts a string with syntax 12:34 or 5:01 into hours and minutes
// no normalisation occurs, values out of bounds result in an error
func decodeClock(input string) (hour, minute int, err error) {
	if !pattern.MatchString(input) {
		err = fmt.Errorf("clock string %s has an unsupported syntax", input)
		return
	}
	submatches := pattern.FindStringSubmatch(input)

	hour, err = strconv.Atoi(submatches[1])
	if err != nil {
		err = fmt.Errorf("error converting hour component of %s; %s", input, err)
		return
	}
	if hour > 23 {
		err = fmt.Errorf("clock hours %v are out of bounds", hour)
		return
	}

	minute, err = strconv.Atoi(submatches[2])
	if err != nil {
		err = fmt.Errorf("error converting minute component of %s; %s", input, err)
		return
	}
	if minute > 59 {
		err = fmt.Errorf("clock minutes %v are out of bounds", minute)
		return
	}
	return
}
