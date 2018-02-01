package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	astro "github.com/kelvins/sunrisesunset"
)

const (
	version = 2
)

var logger *log.Logger
var config = configuration{}

type configuration struct {
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

// sunsetToday returns the time of today's sunset in the system's local time
func sunsetToday(latitude, longitude float64) (time.Time, error) {
	now := time.Now()       // local time now
	_, offset := now.Zone() // offset in seconds

	// GetSunriseSunset expects the UTC in units of hours
	_, sunset, err := astro.GetSunriseSunset(latitude, longitude, float64(offset/3600), now)
	if err != nil {
		return sunset, err
	}

	// the date returned by GetSunriseSunset is the "zero" value so construct a new Time using the current time
	return time.Date(now.Year(), now.Month(), now.Day(), sunset.Hour(), sunset.Minute(), sunset.Second(), 0, now.Location()), nil
}

// disableCache disables the client cache so that a request is sent to the server each and every time
func disableCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

// respond writes the http response and logs the action
func respond(w http.ResponseWriter, msg string, code int) {
	logger.Printf("Response [%v] %v\n", code, msg)
	if code == http.StatusOK {
		fmt.Fprintln(w, msg)
	} else {
		http.Error(w, msg, code)
	}
}

// about reports about the server
func about(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* about request\n")
	fmt.Fprintf(w, "Heihei: version %2d\n", version)
	fmt.Fprintf(w, "        at (%v, %v)\n", config.Latitude, config.Longitude)
}

// sunset reports the time of sunset at the device location
func sunset(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* sunset request\n")
	sunset, err := sunsetToday(config.Latitude, config.Longitude)
	if err != nil {
		respond(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, sunset.Format("Today's (Monday 2 January 2006) sunset is approximately at 15:04:05 MST"), http.StatusOK)
}

// light controls the RF controlled light
func light(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* light request")

	disableCache(w)

	modes, ok := r.URL.Query()["mode"]
	if !ok || len(modes) < 1 {
		respond(w, "Missing 'mode' value", http.StatusUnprocessableEntity)
		return
	}

	mode := modes[0]
	logger.Printf("mode = %v", mode)
	switch mode {
	case "on":
		if err := setPlug(plugOne, true); err != nil {
			respond(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respond(w, "on", http.StatusOK)
	case "off":
		if err := setPlug(plugOne, false); err != nil {
			respond(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respond(w, "off", http.StatusOK)
	default:
		respond(w, fmt.Sprintf("Unknown 'mode' value '%v'", mode), http.StatusUnprocessableEntity)
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
	http.HandleFunc("/sunset", sunset)
	logger.Fatal(http.ListenAndServe(":8000", nil))
}
