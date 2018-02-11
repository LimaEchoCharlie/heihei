package main

import (
	"context"
	"time"
)

type alarm struct {
	setc   chan bool
	ticker *time.Ticker
}

// newAlarm creates a new alarm
func newAlarm(ctx context.Context, accuracy time.Duration) alarm {
	a := alarm{
		setc:   make(chan bool),
		ticker: time.NewTicker(accuracy),
	}

	// start routine
	go func() {
		on := false
		for {
			select {

			case on = <-a.setc:
			case now := <-a.ticker.C:
				if on {
					logger.Printf("%v\n", now)
				}
			case <-ctx.Done():
				a.ticker.Stop()
			}
		}
	}()

	return a
}

func (a alarm) set(on bool) {
	a.setc <- on
}
