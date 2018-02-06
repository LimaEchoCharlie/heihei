package main

import (
	"context"
	"fmt"
	"math/rand"
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
type plugID int

const (
	plugAll plugID = iota
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

// initPlugs initialises the pins used to communicate with the plugs
func initPlugs() (err error) {
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

type plug struct {
	id      plugID
	setChan chan bool
	getChan chan bool
	timer   *time.Timer
}

// newPlug creates a new variable to control the plug with the supplied id
func newPlug(ctx context.Context, id plugID) plug {
	l := plug{
		setChan: make(chan bool),
		getChan: make(chan bool),
		id:      id,
	}
	go l.controller(ctx)
	return l
}

// setPins turns plug on or off by setting the pins directly.
// This function isn't intended to called from outside this file.
func (l plug) setPins(on bool) error {
	// lock pins
	mutex.Lock()
	defer mutex.Unlock()

	// clear error
	clearPinError()

	// set d2-d1-d0 depending on which plug
	switch l.id {
	case plugAll:
		// 011
		d2.off()
		d1.on()
		d0.on()
	case plugOne:
		// 111
		logger.Println("plug one")
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
		return fmt.Errorf("%d is not a valid plug id", l.id)
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

// controller manages the plug in a new go routine
func (l plug) controller(ctx context.Context) {
	// initialise the plugs
	if err := initPlugs(); err != nil {
		logger.Fatal(err)
	}

	// start with plug off
	l.setPins(false)
	currentState := false

	go func() {
		for {
			select {

			case newState := <-l.setChan:
				l.setPins(newState)
				currentState = newState
			case l.getChan <- currentState:
			case <-ctx.Done():
				return
			}
		}
	}()
}

// set sets the plug
func (l plug) set(on bool) {
	logger.Printf("set %v\n", on)
	l.setChan <- on
}

// setForDuration sets the plug to on and reverts to the inverse state at the end of the duration
func (l plug) setForDuration(on bool, d time.Duration) {
	r := rand.Intn(1000)
	logger.Printf("[%03d] setForDuration start\n", r)
	l.set(on)
	f := func() {
		logger.Printf("[%03d] setForDuration finish\n", r)
		l.set(!on)
	}
	l.timer = time.AfterFunc(d, f)
}

// state returns the current status of the plug
func (l plug) state() bool {
	return <-l.getChan
}
