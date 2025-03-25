package main

import (
	"testing"
)

func Test_generateMatchMakerMatchesUntilSuccess(t *testing.T) {
	tests := []struct {
		name    string
	}{
		{
			"TestWithCsv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generateMatchMakerMatchesUntilSuccess()
			if err != nil {
				t.Errorf("generateMatchMakerMatchesUntilSuccess() error = %v", err)
				return
			}
			t.Error("Test actually success, just want to print out logs")
		})
	}
}
