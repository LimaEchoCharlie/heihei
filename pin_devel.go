// +build !rapi

package main

import "log"

const buildType = "devel"

// pin definitions
var (
	// encoder (by board position)
	d0 pin = "d0"
	d1 pin = "d1"
	d2 pin = "d2"
	d3 pin = "d3"
	// modulator mode
	mode pin = "mode"
	// modulator enable
	enable pin = "enable"
)

type pin string

type pinLevel bool

const (
	pinLow  pinLevel = false
	pinHigh pinLevel = true
)

// initHAL initialises the hardware abstraction layer
func initHAL() error {
	return nil
}

// setLevel changes the level of the pin
func setLevel(p pin, l pinLevel) (err error) {
	log.Printf("pin %s set %v\n", p, l)
	return nil
}
