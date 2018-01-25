package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	version = 1
)

type Configuration struct {
	Latitude  float64 // the latitude of the device
	Longitude float64 // the longitude of the device
}

var config = Configuration{}

// loadConfiguration loads the server configuration
func loadConfiguration() error {
	// open the configuration file that is in the same directory as the executable
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(filepath.Dir(ex), "configuration.json"))
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

// about reports about the server
func about(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Heihei: version %2d\n", version)
}

// howdy echoes howdy
func howdy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Howdy, this is Heihei at (%v, %v)\n", config.Latitude, config.Longitude)
}

func main() {
	if err := loadConfiguration(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/howdy", howdy)
	http.HandleFunc("/about", about)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
