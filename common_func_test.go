package main

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
)

func Test_assignMatchesToCourts(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			"ReallocateCourtWithoutChangingMatches",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load from csv and take average of all ranking score. also output average team ranking
			matchesWithWrongCourtAllocCSVFile, err := os.OpenFile("SoloHunter_202503271547.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				t.Error(err)
				return
			}
			defer matchesWithWrongCourtAllocCSVFile.Close()

			matchMetadatas := []MatchMetadata{}
			if err := gocsv.UnmarshalFile(matchesWithWrongCourtAllocCSVFile, &matchMetadatas); err != nil {
				// Load clients from file
				t.Error(err)
				return
			}

			count := 0
			got, err := assignMatchesToCourts(matchMetadatas, []int{6, 7, 8, 9}, true)
			for err != nil && count < 10000000 {
				// Randomly shuffle
				count++
				teamRandomizeOrder := make([]MatchMetadata, len(matchMetadatas))
				copy(teamRandomizeOrder, matchMetadatas)
				for i := range teamRandomizeOrder {
					j := rand.Intn(i + 1)
					teamRandomizeOrder[i], teamRandomizeOrder[j] = teamRandomizeOrder[j], teamRandomizeOrder[i]
				}
				matchMetadatas = teamRandomizeOrder
				got, err = assignMatchesToCourts(matchMetadatas, []int{6, 7, 8, 9}, true)
			}

			if err != nil {
				panic(err)
			}
			
			// Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
			SoloHunterFile, err := os.OpenFile(fmt.Sprintf("SoloHunter_Realloc_Courts_NoShuffle_%s_count_%d.csv", time.Now().Format("200601021504"), count), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				panic(err)
			}
			defer SoloHunterFile.Close()
			if err = gocsv.MarshalFile(&got, SoloHunterFile); err != nil {
				panic(err)
			}
			t.Error("Actually success, just want to see the log")
		})
	}
}
