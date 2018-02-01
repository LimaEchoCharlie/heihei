package main

import "context"

var lightSubscriber = make(chan bool)
var lightRequest = make(chan bool)

// startLightController manages the light in a new go routine
func startLightController(ctx context.Context) {
	go func() {
		// start with light off
		setPlug(plugOne, false)
		currentState := false

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
	lightSubscriber <- on
}

// light returns the current status of the light
func light() bool {
	lightRequest <- true
	return <-lightRequest
}
