package main

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

const (
	low  = gpio.Low
	high = gpio.High
)

// pin definitions
var (
	// encoder (by board position)
	d0 = rpi.P1_11
	d1 = rpi.P1_15
	d2 = rpi.P1_16
	d3 = rpi.P1_13
	// modulator mode
	mode = rpi.P1_18
	// modulator enable
	enable = rpi.P1_22
)

// plug ids
const (
	plugAll = iota
	plugOne
	plugTwo
)

// initPlug initialises the pins used to communicate with the plugs
func initPlug() error {
	// initialise periph
	if _, err := host.Init(); err != nil {
		return err
	}

	// set encoder to 0000
	d3.Out(low)
	d2.Out(low)
	d1.Out(low)
	d0.Out(low)

	// diable modulator
	enable.Out(low)

	// set modulator to ASK
	mode.Out(low)
	return nil
}

// setPlug turns plug (with id) on or off
func setPlug(id int, on bool) error {
	// set d2-d1-d0 depending on which plug
	switch id {
	case plugAll:
		// 011
		d2.Out(low)
		d1.Out(high)
		d0.Out(high)
	case plugOne:
		// 111
		d2.Out(high)
		d1.Out(high)
		d0.Out(high)
	case plugTwo:
		// 110
		d2.Out(high)
		d1.Out(high)
		d0.Out(low)
	default:
		// not recognised, return error
		return fmt.Errorf("%d is not a valid plug id", id)
	}

	// set d3 depending on on/off
	if on {
		d3.Out(high)
	} else {
		d3.Out(low)
	}

	// allow the encoder to settle
	time.Sleep(100 * time.Millisecond)

	// enable the modulator
	enable.Out(high)
	// pause
	time.Sleep(250 * time.Millisecond)
	// disable the modulator
	enable.Out(low)

	return nil
}
