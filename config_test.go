package main

import (
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
