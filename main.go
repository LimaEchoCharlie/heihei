package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	version = 1
)

var logger *log.Logger
var config = Configuration{}

type Configuration struct {
	Latitude  float64 // the latitude of the device
	Longitude float64 // the longitude of the device
}

// initLogger creates a new logger that writes to out
func initLogger(out io.Writer) {
	logger = log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// loadConfiguration loads the server configuration
func loadConfiguration(path string) error {
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

// about reports about the server
func about(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* about request\n")
	fmt.Fprintf(w, "Heihei: version %2d\n", version)
	fmt.Fprintf(w, "        at (%v, %v)\n", config.Latitude, config.Longitude)
}

// light controls the RF controlled light
func light(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* light request\n")

	modes, ok := r.URL.Query()["mode"]
	if !ok || len(modes) < 1 {
		fmt.Fprintf(w, "Url Param 'mode' is missing\n")
		return
	}

	switch modes[0] {
	case "on":
		fmt.Fprintf(w, "on\n")
		setPlug(plugOne, true)
	case "off":
		fmt.Fprintf(w, "off\n")
		setPlug(plugOne, false)
	default:
		fmt.Fprintf(w, "Url Param 'mode' value is unknown\n")
		return
	}
	return
}

func main() {
	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(ex)

	// setup logging
	// If the logfile doesn't exist, create it. Otherwise append to it.
	f, err := os.OpenFile(filepath.Join(path, "heihei.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	initLogger(f)

	logger.Printf("***************\n")
	logger.Printf("starting Heihei\n")

	// load the configuration
	if err := loadConfiguration(path); err != nil {
		logger.Fatal(err)
	}
	// initialise the plugs
	if err := initPlug(); err != nil {
		logger.Fatal(err)
	}

	// register the handlers and listen
	http.HandleFunc("/about", about)
	http.HandleFunc("/light", light)
	logger.Fatal(http.ListenAndServe(":8000", nil))
}
