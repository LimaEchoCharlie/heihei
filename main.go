package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	astro "github.com/kelvins/sunrisesunset"
)

const (
	version = 3
)

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
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* about request\n")
	disableCache(w)
	fmt.Fprintf(w, "Heihei: version %2d\n", version)
	latitude, longitude := location()
	fmt.Fprintf(w, "        at (%v, %v)\n", latitude, longitude)
	fmt.Fprintf(w, "        light is %v\n", light())
	if isDevel() {
		fmt.Fprintf(w, "        DEVEL\n")
	}
}

// sunset reports the time of sunset at the device location
func sunsetHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* sunset request\n")
	latitude, longitude := location()
	sunset, err := sunsetToday(latitude, longitude)
	if err != nil {
		respond(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, sunset.Format("Today's (Monday 2 January 2006) sunset is approximately at 15:04:05 MST"), http.StatusOK)
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
		setLightForDuration(on, d)
	} else {
		respond(w, msg, http.StatusOK)
		setLight(on)
	}
}

// lightHandler controls the RF controlled light
func lightHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("* light request")

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

func main() {
	// initialise logging
	if err := initLogging(); err != nil {
		panic(err)
	}
	defer stopLogging()

	logger.Printf("***************\n")
	logger.Printf("starting Heihei\n")

	// load the configuration
	if err := loadConfiguration(); err != nil {
		logger.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start the light controller
	startLightController(ctx)

	// register the handlers and listen
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/light", lightHandler)
	http.HandleFunc("/sunset", sunsetHandler)
	logger.Fatal(http.ListenAndServe(":8000", nil))
}
