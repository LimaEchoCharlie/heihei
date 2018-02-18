package main

import (
	"log"

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
	devel  string
)

type pin struct {
	gpio.PinIO
}

type pinLevel gpio.Level

const (
	pinLow  pinLevel = pinLevel(gpio.Low)
	pinHigh pinLevel = pinLevel(gpio.High)
)

func isDevel() bool {
	return devel != ""
}

func initHAL() error {
	_, err := host.Init()
	return err
}

// setLevel changes the level of the pin
func setLevel(p pin, l pinLevel) (err error) {
	if isDevel() {
		return nil
	}

	err = p.Out(gpio.Level(l))
	if err != nil {
		log.Printf("%v %v failure %v\n", p, l, err)
	}
	return
}
