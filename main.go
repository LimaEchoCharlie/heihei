package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	version        = 5
	configFilename = "configuration.json"
	logFilename    = "heihei.log"
)

var (
	lightOne plug
	alarmOne alarm
	config   configuration
)

func init() {
	// discard logging by default i.e. for unit testing
	log.SetOutput(ioutil.Discard)
}

// disableCache disables the client cache so that a request is sent to the server each and every time
func disableCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

// respond writes the http response and logs the action
func respond(w http.ResponseWriter, msg string, code int) {
	log.Printf("Response [%v] %v\n", code, msg)
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
	fmt.Fprintf(w, "        build type %s\n", buildType)
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

// getDuration extracts a duration from the query; only positive values are returned
// zero is returned if an error occurs
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
	if secs < 0 {
		secs = 0
	}
	return time.Duration(secs) * time.Second
}

// lightModeHandler deals with light requests when the mode is known
func lightModeHandler(w http.ResponseWriter, r *http.Request, on bool) {
	log.Printf("mode = %v", on)
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

// notifyHandler sets a notification
func notifyHandler(w http.ResponseWriter, r *http.Request) {
	disableCache(w)

	query, ok := r.URL.Query()["time"]
	if !ok || len(query) < 1 {
		respond(w, "Missing 'time' value", http.StatusUnprocessableEntity)
		return
	}

	hour, minute, err := decodeClock(query[0])
	if err != nil {
		respond(w, "Invalid 'time' value", http.StatusUnprocessableEntity)
		return
	}

	timer, err := newNotification(nextTime(time.Now(), hour, minute))
	if err != nil {
		respond(w, "Notification error", http.StatusInternalServerError)
		return
	}

	go func() {
		<-timer.C
		log.Println("notification fired")
	}()

	respond(w, "Notification set", http.StatusOK)
	return
}

// fileHandlerFunc outputs the file with path to the browser
func fileHandlerFunc(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		disableCache(w)
		content, err := ioutil.ReadFile(path)
		if err != nil {
			respond(w, fmt.Sprintf("file at \"%s\": %s", path, err), http.StatusInternalServerError)
		}
		// don't use respond as this will write file to the logfile
		fmt.Fprintf(w, "%s", content)
		return
	}
}

// logHandler is a http handler wrapper for logging
func logHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("-> %v: from %v\n", r.URL.Path, r.RemoteAddr)
			h.ServeHTTP(w, r) // call original
			log.Printf("<- %v\n", r.URL.Path)
		})
}

func main() {
	var err error

	// get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(ex)

	// load the configuration
	configFilePath := filepath.Join(path, configFilename)
	configFile, err := os.Open(configFilePath)
	if err != nil {
		panic(err)
	}

	config, err = getConfiguration(configFile)
	if err != nil {
		panic(err)
	}

	// initialise logging
	logfile := os.Stdout
	var logFilePath string
	if !config.logToStdout {
		logFilePath = filepath.Join(path, logFilename)
		logfile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logfile = os.Stdout
		}
	}
	defer logfile.Close()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logfile)

	log.Printf("***************\n")
	log.Printf("starting Heihei\n")

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
	mux.HandleFunc("/notify", notifyHandler)
	mux.HandleFunc("/logfile", fileHandlerFunc(logFilePath))
	mux.HandleFunc("/config", fileHandlerFunc(configFilePath))
	log.Fatal(http.ListenAndServe(":8000", logHandler(mux)))
}
