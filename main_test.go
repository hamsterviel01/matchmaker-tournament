package main

import (
	"testing"
)

func Test_generateMatchMakerMatches(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"TestWithCSV",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateMatchMakerMatches()
		})
	}
}

func Test_generateMatchMakerMatchesUntilSuccess(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			"TestWithCSV",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generateMatchMakerMatchesUntilSuccess()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateMatchMakerMatchesUntilSuccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Errorf("Actually success, just want to see the print out")
		})
	}
}
