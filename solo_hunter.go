package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

func generateSoloHunterMatchesUntilSuccess() ([]MatchMetadata, error) {
	runNo := 0
	matches, err := generateSoloHunterMatches()
	for err != nil && runNo < SOLO_HUNTER_RERUN {
		matches, err = generateSoloHunterMatches()
		log.Warn(err)
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

func generateSoloHunterMatches() ([]MatchMetadata, error) {
	playerAndRanking, playerGender, err := loadAvgRankingAndGender()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}

	soloHunterMatches := []MatchMetadata{}
	soloHunterPlayerAndNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		soloHunterPlayerAndNoOfMatch[player] = 0
	}
	soloHunterPlayerAndOponentNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		for opponent := range playerAndRanking {
			if opponent != player {
				soloHunterPlayerAndOponentNoOfMatch[generateKey(player, opponent)] = 0
			}
		}
	}
	for player1 := range playerAndRanking {
		for player2 := range playerAndRanking {
			for player3 := range playerAndRanking {
				for player4 := range playerAndRanking {
					if isAllPlayersDifferentAndNoTwoFemaleSameTeam([]string{player1, player2, player3, player4}, playerGender) &&
						percentageDifference(player1, player2, player3, player4, playerAndRanking) < SOLO_HUNTER_MAX_RANK_PERCENTAGE_DIFFERENCE &&
						soloHunterPlayerAndNoOfMatch[player1] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player2] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player3] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player4] < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player3)] < SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player4)] < SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player3)] < SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player4)] < SOLO_HUNTER_MAX_REPEATED_OPPONENT {
						soloHunterMatches = append(soloHunterMatches, MatchMetadata{
							Player1:              player1,
							Player2:              player2,
							Player3:              player3,
							Player4:              player4,
							PercentageDifference: percentageDifference(player1, player2, player3, player4, playerAndRanking),
						})
						soloHunterPlayerAndNoOfMatch[player1] = soloHunterPlayerAndNoOfMatch[player1] + 1
						soloHunterPlayerAndNoOfMatch[player2] = soloHunterPlayerAndNoOfMatch[player2] + 1
						soloHunterPlayerAndNoOfMatch[player3] = soloHunterPlayerAndNoOfMatch[player3] + 1
						soloHunterPlayerAndNoOfMatch[player4] = soloHunterPlayerAndNoOfMatch[player4] + 1
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player3)] = soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player3)] + 1
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player4)] = soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player4)] + 1
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player3)] = soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player3)] + 1
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player4)] = soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player4)] + 1
					}
				}
			}
		}
	}

	// Check if player play enough match
	for player, matchesPlayed := range soloHunterPlayerAndNoOfMatch {
		if matchesPlayed != SOLO_HUNTER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("player %s plays %d matches. This mean you have to relax the requirements to generate enough matching, or something wrong and player play more than they should", player, matchesPlayed)
			return nil, err
		}
	}

	// Assign match to court so that no player have to player 2 match in one round
	soloHunterMatches, err = assignMatchesToCourts(soloHunterMatches)
	if err != nil {
		return nil, err
	}

	// Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
	SoloHunterFile, err := os.OpenFile(fmt.Sprintf("SoloHunter_%s.csv", time.Now().Format("200601021504")), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer SoloHunterFile.Close()
	if err = gocsv.MarshalFile(&soloHunterMatches, SoloHunterFile); err != nil {
		panic(err)
	}

	return soloHunterMatches, nil
}
