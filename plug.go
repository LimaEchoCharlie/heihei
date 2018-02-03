package main

import (
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
)

// pin definitions
var (
	// encoder (by board position)
	d0 = pin{rpi.P1_11}
	d1 = pin{rpi.P1_15}
	d2 = pin{rpi.P1_16}
	d3 = pin{rpi.P1_13}
	// modulator mode
	mode = pin{rpi.P1_18}
	// modulator enable
	enable = pin{rpi.P1_22}
)

// mutex to protect pin writes
var mutex = &sync.Mutex{}

// save last error
var pinError error

// plug ids
const (
	plugAll = iota
	plugOne
	plugTwo
)

type pin struct {
	gpio.PinIO
}

// clearPinError clears the saved error related to pin operations
func clearPinError() {
	pinError = nil
}

// setPinError saves not nil errors
func setPinError(err error) {
	if err != nil {
		pinError = err
	}
	return
}

// lastPinError returns the last error from a pin operation
func lastPinError() error {
	return pinError
}

// out changes the level of the pin
func (p pin) out(l gpio.Level) (err error) {
	if isDevel() {
		return nil
	}

	err = p.Out(l)
	setPinError(err)
	if err != nil {
		logger.Printf("%v %v failure %v\n", p, l, err)
	}
	return
}

// off switches off the pin
func (p pin) off() (err error) {
	return p.out(gpio.Low)
}

// on switches on the pin
func (p pin) on() (err error) {
	return p.out(gpio.High)
}

// initPlug initialises the pins used to communicate with the plugs
func initPlug() (err error) {
	// lock mutex
	mutex.Lock()
	defer mutex.Unlock()

	// clear error
	clearPinError()

	// initialise periph
	if _, err := host.Init(); err != nil {
		return err
	}

	// set encoder to 0000
	d3.off()
	d2.off()
	d1.off()
	d0.off()

	// diable modulator
	enable.off()

	// set modulator to ASK
	mode.off()
	return lastPinError()
}

// setPlug turns plug (with id) on or off
func setPlug(id int, on bool) error {
	// lock mutex
	mutex.Lock()
	defer mutex.Unlock()

	// clear error
	clearPinError()

	// set d2-d1-d0 depending on which plug
	switch id {
	case plugAll:
		// 011
		d2.off()
		d1.on()
		d0.on()
	case plugOne:
		// 111
		d2.on()
		d1.on()
		d0.on()
	case plugTwo:
		// 110
		d2.on()
		d1.on()
		d0.off()
	default:
		// not recognised, return error
		return fmt.Errorf("%d is not a valid plug id", id)
	}

	// set d3 depending on on/off
	if on {
		d3.on()
	} else {
		d3.off()
	}

	// allow the encoder to settle
	time.Sleep(100 * time.Millisecond)

	// enable the modulator
	enable.on()
	// pause
	time.Sleep(250 * time.Millisecond)
	// disable the modulator
	enable.off()

	return lastPinError()
}
