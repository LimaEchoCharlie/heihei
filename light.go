package main

import (
	"context"
	"math/rand"
	"time"
)

var (
	lightSubscriber = make(chan bool)
	lightRequest    = make(chan bool)
	timer           *time.Timer
)

// startLightController manages the light in a new go routine
func startLightController(ctx context.Context) {
	// initialise the plugs
	if err := initPlug(); err != nil {
		logger.Fatal(err)
	}

	// start with light off
	setPlug(plugOne, false)
	currentState := false

	go func() {
		for {
			select {

			case newState := <-lightSubscriber:
				setPlug(plugOne, newState)
				currentState = newState
			case <-lightRequest:
				lightRequest <- currentState
			case <-ctx.Done():
				return
			}
		}
	}()
}

// setLight sets the light
func setLight(on bool) {
	logger.Printf("setLight %v\n", on)
	lightSubscriber <- on
}

// setLightFirDuration sets the light to on and reverts to the inverse state at the end of the duration
func setLightForDuration(on bool, d time.Duration) {
	r := rand.Intn(1000)
	logger.Printf("[%03d] setLightForDuration start\n", r)
	setLight(on)
	f := func() {
		logger.Printf("[%03d] setLightForDuration finish\n", r)
		setLight(!on)
	}
	timer = time.AfterFunc(d, f)
}

// light returns the current status of the light
func light() bool {
	lightRequest <- true
	return <-lightRequest
}
