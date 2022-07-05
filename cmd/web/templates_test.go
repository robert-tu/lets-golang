package main

import (
	"testing"
	"time"
)

func TestHumanDate(t *testing.T) {
	// create slice of anonymous structs
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2021, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "Dec 17 2021 at 10:00",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2020, 12, 31, 0, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "Dec 30 2020 at 23:00",
		},
	}

	// loop test cases
	for _, tt := range tests {
		// run each sub test
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			if hd != tt.want {
				t.Errorf("want %q; got %q", tt.want, hd)
			}
		})
	}
}
