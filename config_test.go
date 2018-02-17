package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeTimeError(t *testing.T) {
	testCases := []struct {
		input string
	}{
		{"-10"},
		{"-1:10"},
		{"25:10"},
		{"5:70"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input %v", tc.input), func(t *testing.T) {

			if _, _, err := decodeClock(tc.input); err == nil {
				t.Errorf("expected error for input %v; but got none", tc.input)
			}
		})
	}
}

func TestDecodeTimeValid(t *testing.T) {
	testCases := []struct {
		input        string
		hour, minute int
	}{
		{"01:10", 1, 10},
		{"5:59", 5, 59},
		{"12:00", 12, 0},
		{"23:43", 23, 43},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input %v", tc.input), func(t *testing.T) {

			hour, minute, err := decodeClock(tc.input)

			if err != nil {
				t.Errorf("unexpected error for input %v; %v", tc.input, err)
			} else if hour != tc.hour {
				t.Errorf("incorrect hour value %v; expected %v", hour, tc.hour)
			} else if minute != tc.minute {
				t.Errorf("incorrect minute value %v; expected %v", minute, tc.minute)
			}

		})
	}
}

const (
	magNLat = 80.31
	magNLon = -72.62
	bedtime = "23:34"
)

var validConfig = fmt.Sprintf(`{"location":[%f, %f], "lights_out":"%s", "log_to_stdout":true}`,
	magNLat, magNLon, bedtime)

func TestGetConfigError(t *testing.T) {

	testCases := []struct {
		raw  *bytes.Buffer
		note string
	}{
		{raw: bytes.NewBufferString(`{}`), note: "empty json string"},
		{raw: bytes.NewBufferString(fmt.Sprintf(`{"lights_out":"%s"}`, bedtime)), note: "missing location"},
		{raw: bytes.NewBufferString(fmt.Sprintf(`{"location":[%f, %f]}`, magNLat, magNLon)), note: "missing lightsOut"},
		{raw: bytes.NewBufferString(fmt.Sprintf(`{"location":[%f], "lights_out":"%s"}`, magNLat, bedtime)),
			note: "location too short"},
		{raw: bytes.NewBufferString(fmt.Sprintf(`{"location":[%f,%f,%f], "lights_out":"%s"}`, magNLat, magNLon, 1.2, bedtime)),
			note: "location too long"},
		{raw: bytes.NewBufferString(fmt.Sprintf(`{"location":[%f,%f], "lights_out":"%s"}`, magNLat, magNLon, "111:78")),
			note: "invalid clock time"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input %v", tc.note), func(t *testing.T) {
			if _, err := getConfiguration(tc.raw); err == nil {
				t.Errorf("expected error for input %v; but got none", tc.note)
			}
		})
	}
}

func TestGetConfigLocation(t *testing.T) {
	buf := bytes.NewBufferString(validConfig)
	config, err := getConfiguration(buf)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	lat, lon := config.latLong()
	if lat != magNLat {
		t.Errorf("Got latitude %v; expected %v", lat, magNLat)
	}
	if lon != magNLon {
		t.Errorf("Got longtitude %v; expected %v", lon, magNLon)
	}
}

func TestGetConfigLightsOut(t *testing.T) {
	buf := bytes.NewBufferString(validConfig)
	config, err := getConfiguration(buf)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if config.lightsOut != bedtime {
		t.Errorf("Got lights out at %v; but bedtime is %v", config.lightsOut, bedtime)
	}
}

func TestGetConfigLogToStdout(t *testing.T) {
	buf := bytes.NewBufferString(validConfig)
	config, err := getConfiguration(buf)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if !config.logToStdout {
		t.Errorf("log to stdout is false; expected true")
	}
}
