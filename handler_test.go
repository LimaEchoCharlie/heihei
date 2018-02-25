package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetDuration(t *testing.T) {
	testCases := []struct {
		values   string
		expected int // expected in seconds
		note     string
	}{
		{"", 0, "empty"},
		{"?spam=2", 0, "wrong query"},
		{"?secs=eggs", 0, "strign value"},
		{"?secs=2.6", 0, "float value"},
		{"?secs=-2", 0, "negative value"},
		{"?secs=2", 2, "correct"},
	}
	for _, tc := range testCases {
		t.Run(tc.note, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/test-request%s", tc.values), nil)
			expect := time.Duration(tc.expected) * time.Second
			if d := getDuration(req); d != expect {
				t.Errorf("duration %s: got %v want %v", req.URL, d, expect)
			}
		})
	}
}
