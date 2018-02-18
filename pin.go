package main

// save last error
var pinError error

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

// off switches off the pin
func (p pin) off() (err error) {
	err = setLevel(p, pinLow)
	setPinError(err)
	return
}

// on switches on the pin
func (p pin) on() (err error) {
	err = setLevel(p, pinHigh)
	setPinError(err)
	return
}
