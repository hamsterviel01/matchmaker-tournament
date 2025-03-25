package main

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

type MatchMakerMatch struct {
	Court   int    `csv:"court"`
	Player1 string `csv:"player1"`
	Player2 string `csv:"player2"`
	Player3 string `csv:"player3"`
	Player4 string `csv:"player4"`
}

func generateMatchMakerMatchesUntilSuccess() ([]MatchMakerMatch, error) {
	runNo := 0
	matches, err := generateMatchMakerMatches()
	for err != nil && runNo < MATCH_MAKER_RERUN {
		matches, err = generateMatchMakerMatches()
		runNo++
	}
	if err != nil {
		return nil, err
	}
	for _, match := range matches {
		log.Infof("%s,%s,%s,%s", match.Player1, match.Player2, match.Player3, match.Player4)
	}
	log.Infof("successfully generate matches after %d runs", runNo)
	return matches, nil
}

func generateMatchMakerMatches() ([]MatchMakerMatch, error) {
	playerAndRanking, err := loadAvgRanking()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}

	MatchMakerMatches := []MatchMakerMatch{}
	MatchMakerPlayerAndNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		MatchMakerPlayerAndNoOfMatch[player] = 0
	}
	MatchMakerPlayerAndOponentNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		for opponent := range playerAndRanking {
			if opponent != player {
				MatchMakerPlayerAndOponentNoOfMatch[generateKey(player, opponent)] = 0
			}
		}
	}
	for player1 := range playerAndRanking {
		for player2 := range playerAndRanking {
			for player3 := range playerAndRanking {
				for player4 := range playerAndRanking {
					if isAllPlayersDifferent([]string{player1, player2, player3, player4}) &&
						percentageDifference(player1, player2, player3, player4, playerAndRanking) < MATCH_MAKER_MAX_RANK_PERCENTAGE_DIFFERENCE &&
						MatchMakerPlayerAndNoOfMatch[player1] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						MatchMakerPlayerAndNoOfMatch[player2] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						MatchMakerPlayerAndNoOfMatch[player3] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						MatchMakerPlayerAndNoOfMatch[player4] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] < MATCH_MAKER_MAX_REPEATED_OPPONENT {
						MatchMakerMatches = append(MatchMakerMatches, MatchMakerMatch{
							Player1: player1,
							Player2: player2,
							Player3: player3,
							Player4: player4,
						})
						MatchMakerPlayerAndNoOfMatch[player1] = MatchMakerPlayerAndNoOfMatch[player1] + 1
						MatchMakerPlayerAndNoOfMatch[player2] = MatchMakerPlayerAndNoOfMatch[player2] + 1
						MatchMakerPlayerAndNoOfMatch[player3] = MatchMakerPlayerAndNoOfMatch[player3] + 1
						MatchMakerPlayerAndNoOfMatch[player4] = MatchMakerPlayerAndNoOfMatch[player4] + 1
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] = MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] + 1
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] = MatchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] + 1
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] = MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] + 1
						MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] = MatchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] + 1
					}
				}
			}
		}
	}

	// Check if player play enough match
	for player, matchesPlayed := range MatchMakerPlayerAndNoOfMatch {
		if matchesPlayed != MATCH_MAKER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("player %s plays %d matches. This mean you have to relax the requirements to generate enough matching, or check why player has to play more than they should", player, matchesPlayed)
			return nil, err
		}
	}

	// Randomly shuffle the matches order
	for i := range MatchMakerMatches {
		j := rand.Intn(i + 1)
		MatchMakerMatches[i], MatchMakerMatches[j] = MatchMakerMatches[j], MatchMakerMatches[i]
	}

	// Assign match to court so that no player have to player 2 match in one round
	playersInCurrentRound := []string{}
	for i := range MatchMakerMatches {
		courtNo := (i+1)%NUMBER_OF_COURT
		if courtNo == 0 {
			courtNo = 4
		}

		// If any of 4 player already play this round, find closest group of 4 players that hasn't play and swap the index
		if slices.Contains(playersInCurrentRound, MatchMakerMatches[i].Player1) ||
			slices.Contains(playersInCurrentRound, MatchMakerMatches[i].Player2) ||
			slices.Contains(playersInCurrentRound, MatchMakerMatches[i].Player3) ||
			slices.Contains(playersInCurrentRound, MatchMakerMatches[i].Player4) {
			foundMatch := false
			for j := i + 1; j < len(MatchMakerMatches); j++ {
				if !slices.Contains(playersInCurrentRound, MatchMakerMatches[j].Player1) &&
					!slices.Contains(playersInCurrentRound, MatchMakerMatches[j].Player2) &&
					!slices.Contains(playersInCurrentRound, MatchMakerMatches[j].Player3) &&
					!slices.Contains(playersInCurrentRound, MatchMakerMatches[j].Player4) {
					MatchMakerMatches[i], MatchMakerMatches[j] = MatchMakerMatches[j], MatchMakerMatches[i]
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				return nil, fmt.Errorf("cannot allocate court for some reason, playersInCurrentRound = %v, remaining matches = %v", playersInCurrentRound, MatchMakerMatches[i+1:])
			}
		}

		MatchMakerMatches[i].Court = courtNo
		playersInCurrentRound = append(playersInCurrentRound, MatchMakerMatches[i].Player1, MatchMakerMatches[i].Player2, MatchMakerMatches[i].Player3, MatchMakerMatches[i].Player4)
	}

	// Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
	MatchMakerFile, err := os.OpenFile(fmt.Sprintf("MatchMaker_%s.csv", time.Now().Format("200601021504")), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer MatchMakerFile.Close()
	if err = gocsv.MarshalFile(&MatchMakerMatches, MatchMakerFile); err != nil {
		panic(err)
	}

	return MatchMakerMatches, nil
}
