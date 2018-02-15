package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

type configuration struct {
	location  [2]float64 // the [latitude, longitude] of the device
	lightsOut string
}

// latLong returns the latitude and longitude of the device
func (c configuration) latLong() (float64, float64) {
	return c.location[0], c.location[1]
}

// getConfiguration extracts the server configuration
func getConfiguration(file io.Reader) (config configuration, err error) {

	// use pointers for required values
	ptrConfig := struct {
		Location  *[2]float64 `json:"location"`
		LightsOut *string     `json:"lights_out"`
	}{}
	// decode json
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ptrConfig)
	if err != nil {
		return
	}

	if ptrConfig.Location == nil {
		err = fmt.Errorf("Location is missing from configuration")
		return
	}

	// check that a lights out value has been supplied and that it has a valid syntax
	if ptrConfig.LightsOut == nil {
		err = fmt.Errorf("Lights out is missing from configuration")
		return
	} else if _, _, err = decodeClock(*ptrConfig.LightsOut); err != nil {
		err = fmt.Errorf("Lights out value from configuration decoding error; %s", err)
		return
	}

	config.location = *ptrConfig.Location
	config.lightsOut = *ptrConfig.LightsOut

	return
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
