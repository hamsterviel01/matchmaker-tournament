package main

import (
	"testing"
)

func Test_generateSoloHunterMatchesUntilSuccess(t *testing.T) {
	tests := []struct {
		name    string
	}{
		{
			"TestWithCSV",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generateSoloHunterMatchesUntilSuccess()
			if err != nil {
				t.Errorf("generateSoloHunterMatchesUntilSuccess() error = %v", err)
				return
			}
			t.Error("Actually success, just want to see the print out")
		})
	}
}
