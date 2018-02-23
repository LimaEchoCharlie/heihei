package main

import (
	"testing"
	"time"
)

func TestNextTime(t *testing.T) {
	input := time.Date(2010, 4, 15, 12, 26, 34, 0, time.UTC)
	expected := time.Date(2010, 4, 15, 22, 30, 0, 0, time.UTC)
	actual := nextTime(input, 22, 30)
	if !actual.Equal(expected) {
		t.Errorf("%v =/ %v\n", actual, expected)
	}
}

func TestNextTimePlusDay(t *testing.T) {
	input := time.Date(2010, 6, 30, 23, 26, 48, 4, time.UTC)
	expected := time.Date(2010, 7, 1, 21, 0, 0, 0, time.UTC)
	actual := nextTime(input, 21, 0)
	if !actual.Equal(expected) {
		t.Errorf("%v =/ %v\n", actual, expected)
	}
}

func TestNotificationError( t *testing.T){
    _, err := newNotification(time.Now().Add(-time.Second))
    if err == nil {
            t.Errorf("expected err from previous time")
    }
}

func TestNotification( t *testing.T){
    d := 20 * time.Millisecond
    timer, err := newNotification(time.Now().Add(d))
    if err != nil {
        t.Errorf("unexpected err %v", err)
    }
    select {
    case <-timer.C:
    case <-time.After( 2 * d ):
        t.Errorf("timer should of fired by now")
    }
}
