package main

import (
	"context"
	"fmt"
	"log"
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
//go:generate stringer -type=plugID
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
		log.Printf("%v %v failure %v\n", p, l, err)
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
	p := plug{
		setChan: make(chan bool),
		getChan: make(chan bool),
		id:      id,
	}

	// initialise the plugs
	if err := initPlugs(); err != nil {
		log.Fatal(err)
	}

	// start with plug off
	p.setPins(false)
	currentState := false

	// start routine to control and manage plug
	go func() {
		for {
			select {

			case newState := <-p.setChan:
				log.Printf("set %v %v\n", p.id, newState)
				p.setPins(newState)
				currentState = newState
			case p.getChan <- currentState:
			case <-ctx.Done():
				close(p.getChan)
				close(p.setChan)
				if p.timer != nil {
					p.timer.Stop()
				}
				return
			}
		}
	}()

	return p
}

// setPins turns plug on or off by setting the pins directly.
// This function isn't intended to called from outside this file.
func (p *plug) setPins(on bool) error {
	// lock pins
	mutex.Lock()
	defer mutex.Unlock()

	// clear error
	clearPinError()

	// set d2-d1-d0 depending on which plug
	switch p.id {
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
		return fmt.Errorf("%d is not a valid plug id", p.id)
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

// set sets the plug
func (p *plug) set(on bool) {
	p.setChan <- on
}

// setForDuration sets the plug to on and reverts to the inverse state at the end of the duration
func (p *plug) setForDuration(on bool, d time.Duration) {
	r := rand.Intn(1000)
	log.Printf("[%03d] setForDuration start\n", r)
	if p.timer != nil && p.timer.Stop() {
		log.Printf("[%03d] Stopped existing timer\n", r)
	}
	p.set(on)
	f := func() {
		log.Printf("[%03d] setForDuration finish\n", r)
		p.set(!on)
	}
	p.timer = time.AfterFunc(d, f)
}

// state returns the current status of the plug
func (p *plug) state() bool {
	return <-p.getChan
}
