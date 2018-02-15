package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	version        = 4
	configFilename = "configuration.json"
)

var (
	lightOne plug
	alarmOne alarm
	config   configuration
)

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
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	fmt.Fprintf(w, "Heihei: version %2d\n", version)
	latitude, longitude := config.latLong()
	fmt.Fprintf(w, "        at (%v, %v)\n", latitude, longitude)
	fmt.Fprintf(w, "        light is %v\n", lightOne.state())
	fmt.Fprintf(w, "        alarm is %v\n", alarmOne.isSet())
	if isDevel() {
		fmt.Fprintf(w, "        DEVEL\n")
	}
}

// sunsetFormatter is a utility function to format a sunset
func sunsetFormatter(when string, sunset time.Time) string {
	return fmt.Sprintf("%s %s", when, sunset.Format("(Monday 2 January 2006) sunset is approximately at 15:04:05 MST"))
}

// sunset reports the time of sunset at the device location
func sunsetHandler(w http.ResponseWriter, r *http.Request) {
	latitude, longitude := config.latLong()
	var err error
	yesterday, err := sunset(latitude, longitude, -1)
	if err != nil {
		respond(w, err.Error(), http.StatusInternalServerError)
		return
	}
	today, err := sunset(latitude, longitude, 0)
	if err != nil {
		respond(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tomorrow, err := sunset(latitude, longitude, 1)
	if err != nil {
		respond(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("%s\n%s\n%s",
		sunsetFormatter("Yeserday's", yesterday),
		sunsetFormatter("Today's", today),
		sunsetFormatter("Tomorrow's", tomorrow))
	respond(w, msg, http.StatusOK)
}

// getDuration extracts a duration from the query
func getDuration(r *http.Request) (d time.Duration) {
	var secs int
	vals, ok := r.URL.Query()["secs"]
	if !ok || len(vals) == 0 {
		return 0
	}
	secs, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// lightModeHandler deals with light requests when the mode is known
func lightModeHandler(w http.ResponseWriter, r *http.Request, on bool) {
	logger.Printf("mode = %v", on)
	msg := "off"
	if on {
		msg = "on"
	}
	if d := getDuration(r); d > 0 {
		respond(w, fmt.Sprintf("%v for %v", d, msg), http.StatusOK)
		lightOne.setForDuration(on, d)
	} else {
		respond(w, msg, http.StatusOK)
		lightOne.set(on)
	}
}

// lightHandler controls the RF controlled light
func lightHandler(w http.ResponseWriter, r *http.Request) {
	disableCache(w)

	modes, ok := r.URL.Query()["mode"]
	if !ok || len(modes) < 1 {
		respond(w, "Missing 'mode' value", http.StatusUnprocessableEntity)
		return
	}

	switch modes[0] {
	case "on":
		lightModeHandler(w, r, true)
	case "off":
		lightModeHandler(w, r, false)
	default:
		respond(w, fmt.Sprintf("Unknown 'mode' value '%v'", mode), http.StatusUnprocessableEntity)
		return
	}
	return
}

// alarmHandler controls the alarm
func alarmHandler(w http.ResponseWriter, r *http.Request) {
	disableCache(w)

	query, ok := r.URL.Query()["set"]
	if !ok || len(query) < 1 {
		respond(w, "Missing 'set' value", http.StatusUnprocessableEntity)
		return
	}

	switch query[0] {
	case "on":
		alarmOne.set(true)
		respond(w, "Alarm set", http.StatusOK)
	case "off":
		alarmOne.set(false)
		respond(w, "Alarm unset", http.StatusOK)
	default:
		respond(w, fmt.Sprintf("Unknown 'set' value '%v'", query[0]), http.StatusUnprocessableEntity)
		return
	}
	return
}

// logHandler is a http handler wrapper for logging
func logHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("-> %v: from %v\n", r.URL.Path, r.RemoteAddr)
			h.ServeHTTP(w, r) // call original
			logger.Printf("<- %v\n", r.URL.Path)
		})
}

func main() {
	var err error

	// initialise logging
	if err := initLogging(); err != nil {
		panic(err)
	}
	defer stopLogging()

	logger.Printf("***************\n")
	logger.Printf("starting Heihei\n")

	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(ex)

	configFile, err := os.Open(filepath.Join(path, configFilename))
	if err != nil {
		panic(err)
	}

	// load the configuration
	config, err = getConfiguration(configFile)
	if err != nil {
		logger.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the light controller
	lightOne = newPlug(ctx, plugOne)

	// create an alarm
	alarmOne = newAlarm(ctx, time.Minute)

	// register the handlers and listen
	mux := http.NewServeMux()
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/light", lightHandler)
	mux.HandleFunc("/alarm", alarmHandler)
	mux.HandleFunc("/sunset", sunsetHandler)
	logger.Fatal(http.ListenAndServe(":8000", logHandler(mux)))
}
