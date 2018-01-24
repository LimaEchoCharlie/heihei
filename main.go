package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Configuration struct {
	Latitude  float64 // the latitude of the device
	Longitude float64 // the longitude of the device
}

var config = Configuration{}

// loadConfiguration loads the server configuration
func loadConfiguration() error {
	// open configuration file
	// TODO: remove the need to hardcode location for systemd
	file, err := os.Open("/home/pi/configuration.json")
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

// howdy echoes howdy
func howdy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Howdy, this is Heihei at %v and %v\n", config.Latitude, config.Longitude)
}

func main() {
	if err := loadConfiguration(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/howdy", howdy)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
