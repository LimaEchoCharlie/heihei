package main

import (
	"context"
	"time"

	astro "github.com/kelvins/sunrisesunset"
)

type alarm struct {
	setC   chan bool
	isSetC chan bool
	ticker *time.Ticker
}

// newAlarm creates a new alarm
func newAlarm(ctx context.Context, accuracy time.Duration) alarm {
	a := alarm{
		setC:   make(chan bool),
		isSetC: make(chan bool),
		ticker: time.NewTicker(accuracy),
	}

	// start routine
	go func() {
		on := false
		for {
			select {

			case on = <-a.setC:
			case a.isSetC <- on:
			case now := <-a.ticker.C:
				if on {
					logger.Printf("%v\n", now)
				}
			case <-ctx.Done():
				close(a.setC)
				close(a.isSetC)
				a.ticker.Stop()
			}
		}
	}()

	return a
}

// set sets the alarm
func (a alarm) set(on bool) {
	a.setC <- on
}

// isSet
func (a alarm) isSet() bool {
	return <-a.isSetC
}

// sunset returns the time of sunset in dayOffset days from today in the system's local time
func sunset(latitude, longitude float64, dayOffset int) (time.Time, error) {
	now := time.Now().Add(time.Duration(dayOffset*24) * time.Hour) // local time plus offset
	_, offset := now.Zone()                                        // offset in seconds

	// GetSunriseSunset expects the UTC in units of hours
	_, sunset, err := astro.GetSunriseSunset(latitude, longitude, float64(offset/3600), now)
	if err != nil {
		return sunset, err
	}

	// the date returned by GetSunriseSunset is the "zero" value so construct a new Time using the current time
	return time.Date(now.Year(), now.Month(), now.Day(), sunset.Hour(), sunset.Minute(), sunset.Second(), 0, now.Location()), nil
}

// nextTime returns the first time at hour:minute after the given day i.e.
// 		if day is earlier than hour:minute than a time on that day is returned
// 		otherwise, a time on the next day is returned
func nextTime(baseTime time.Time, hour, minute int) (new time.Time) {
	new = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), hour, minute, 0, 0, baseTime.Location())

	if new.After(baseTime) {
		return
	}

	// does Date normalise? Yes
	return new.AddDate(0, 0, 1)
}
