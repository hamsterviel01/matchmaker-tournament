package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

func generateMatchMakerMatchesUntilSuccess() ([]MatchMetadata, error) {
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

func generateMatchMakerMatches() ([]MatchMetadata, error) {
	playerAndRanking, playerGender, err := loadAvgRankingAndGender()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}

	matchMakerMatches := []MatchMetadata{}
	matchMakerPlayerAndNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		matchMakerPlayerAndNoOfMatch[player] = 0
	}
	matchMakerPlayerAndOponentNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		for opponent := range playerAndRanking {
			if opponent != player {
				matchMakerPlayerAndOponentNoOfMatch[generateKey(player, opponent)] = 0
			}
		}
	}
	for player1 := range playerAndRanking {
		for player2 := range playerAndRanking {
			for player3 := range playerAndRanking {
				for player4 := range playerAndRanking {
					if isAllPlayersDifferentAndNoTwoFemaleSameTeam([]string{player1, player2, player3, player4}, playerGender) &&
						percentageDifference(player1, player2, player3, player4, playerAndRanking) < MATCH_MAKER_MAX_RANK_PERCENTAGE_DIFFERENCE &&
						matchMakerPlayerAndNoOfMatch[player1] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player2] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player3] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player4] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] < MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] < MATCH_MAKER_MAX_REPEATED_OPPONENT {
						matchMakerMatches = append(matchMakerMatches, MatchMetadata{
							Player1: player1,
							Player2: player2,
							Player3: player3,
							Player4: player4,
						})
						matchMakerPlayerAndNoOfMatch[player1] = matchMakerPlayerAndNoOfMatch[player1] + 1
						matchMakerPlayerAndNoOfMatch[player2] = matchMakerPlayerAndNoOfMatch[player2] + 1
						matchMakerPlayerAndNoOfMatch[player3] = matchMakerPlayerAndNoOfMatch[player3] + 1
						matchMakerPlayerAndNoOfMatch[player4] = matchMakerPlayerAndNoOfMatch[player4] + 1
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] = matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] + 1
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] = matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] + 1
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] = matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] + 1
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] = matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] + 1
					}
				}
			}
		}
	}

	// Check if player play enough match
	for player, matchesPlayed := range matchMakerPlayerAndNoOfMatch {
		if matchesPlayed != MATCH_MAKER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("player %s plays %d matches. This mean you have to relax the requirements to generate enough matching, or check why player has to play more than they should", player, matchesPlayed)
			return nil, err
		}
	}

	// Randomly shuffle the matches order
	for i := range matchMakerMatches {
		j := rand.Intn(i + 1)
		matchMakerMatches[i], matchMakerMatches[j] = matchMakerMatches[j], matchMakerMatches[i]
	}

	// Assign match to court so that no player have to player 2 match in one round
	playersInCurrentRound := []string{}
	for i := range matchMakerMatches {
		courtNo := (i + 1) % NUMBER_OF_COURT
		if courtNo == 0 {
			courtNo = 4
		}

		// If any of 4 player already play this round, find closest group of 4 players that hasn't play and swap the index
		if isPlayerExistInList(playersInCurrentRound, matchMakerMatches[i]) {
			foundMatch := false
			for j := i + 1; j < len(matchMakerMatches); j++ {
				if !isPlayerExistInList(playersInCurrentRound, matchMakerMatches[i]) {
					matchMakerMatches[i], matchMakerMatches[j] = matchMakerMatches[j], matchMakerMatches[i]
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				return nil, fmt.Errorf("cannot allocate court for some reason, playersInCurrentRound = %v, remaining matches = %v", playersInCurrentRound, matchMakerMatches[i+1:])
			}
		}

		matchMakerMatches[i].Court = courtNo
		playersInCurrentRound = append(playersInCurrentRound, matchMakerMatches[i].Player1, matchMakerMatches[i].Player2, matchMakerMatches[i].Player3, matchMakerMatches[i].Player4)
	}

	// Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
	MatchMakerFile, err := os.OpenFile(fmt.Sprintf("MatchMaker_%s.csv", time.Now().Format("200601021504")), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer MatchMakerFile.Close()
	if err = gocsv.MarshalFile(&matchMakerMatches, MatchMakerFile); err != nil {
		panic(err)
	}

	return matchMakerMatches, nil
}
