package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

// blink turns the plug off and on
func blink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am going to blink now\n")
	var msg string
	on := true
	for i := 0; i < 10; i++ {
		if on {
			msg = "on"
		} else {
			msg = "off"
		}
		fmt.Fprintf(w, "\t%v\n", msg)
		setPlug(plugOne, on)
		on = !on
		time.Sleep(10 * time.Second)
	}
	fmt.Fprintf(w, "gorffenedig\n")
}

func main() {
	// load the configuration
	if err := loadConfiguration(); err != nil {
		log.Fatal(err)
	}
	// initialise the plugs
	if err := initPlug(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/howdy", howdy)
	http.HandleFunc("/about", about)
	http.HandleFunc("/blink", blink)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
