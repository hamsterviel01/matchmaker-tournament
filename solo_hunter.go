package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

type SoloHunterMatch struct {
	Court   int    `csv:"court"`
	Player1 string `csv:"player1"`
	Player2 string `csv:"player2"`
	Player3 string `csv:"player3"`
	Player4 string `csv:"player4"`
}

func generateSoloHunterMatchesUntilSuccess() ([]SoloHunterMatch, error) {
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

func generateSoloHunterMatches() ([]SoloHunterMatch, error) {
	playerAndRanking, err := loadAvgRanking()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}

	soloHunterMatches := []SoloHunterMatch{}
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
					if isAllPlayersDifferent([]string{player1, player2, player3, player4}) &&
						math.Abs(playerAndRanking[player1]+playerAndRanking[player2]-playerAndRanking[player3]-playerAndRanking[player4]) < SOLO_HUNTER_MAX_RANK_DIFFERENCE &&
						soloHunterPlayerAndNoOfMatch[player1] <= SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player2] <= SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player3] <= SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndNoOfMatch[player4] <= SOLO_HUNTER_TOTAL_MATCH_PER_PERSON &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player3)] <= SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player1, player4)] <= SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player3)] <= SOLO_HUNTER_MAX_REPEATED_OPPONENT &&
						soloHunterPlayerAndOponentNoOfMatch[generateKey(player2, player4)] <= SOLO_HUNTER_MAX_REPEATED_OPPONENT {
						soloHunterMatches = append(soloHunterMatches, SoloHunterMatch{
							Player1: player1,
							Player2: player2,
							Player3: player3,
							Player4: player4,
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

	// Randomly shuffle the matches order
	for i := range soloHunterMatches {
		j := rand.Intn(i + 1)
		soloHunterMatches[i], soloHunterMatches[j] = soloHunterMatches[j], soloHunterMatches[i]
	}

	// Assign match to court so that no player have to player 2 match in one round
	playersInCurrentRound := []string{}
	for i := range soloHunterMatches {
		if (i+1)%NUMBER_OF_COURT == 0 {
			playersInCurrentRound = []string{}
		}

		// If any of 4 player already play this round, find closest group of 4 players that hasn't play and swap the index
		if isPlayerExistInList(playersInCurrentRound, soloHunterMatches[i]) {
			foundMatch := false
			for j := i + 1; j < len(soloHunterMatches); j++ {
				if !isPlayerExistInList(playersInCurrentRound, soloHunterMatches[j]) {
					soloHunterMatches[i], soloHunterMatches[j] = soloHunterMatches[j], soloHunterMatches[i]
					foundMatch = true
					break
				}
			}
			if !foundMatch && i < len(soloHunterMatches) - 3 {
				err = fmt.Errorf("cannot allocate court for some reason, playersInCurrentRound = %v, remaining matches = %v", playersInCurrentRound, soloHunterMatches[i+1:])
				log.Error(err)
				return nil, err
			}
		}

		soloHunterMatches[i].Court = i%NUMBER_OF_COURT + 1
		playersInCurrentRound = append(playersInCurrentRound, soloHunterMatches[i].Player1, soloHunterMatches[i].Player2, soloHunterMatches[i].Player3, soloHunterMatches[i].Player4)
	}

	// Check if player play enough match
	for player, matchesPlayed := range soloHunterPlayerAndNoOfMatch {
		if matchesPlayed < SOLO_HUNTER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("player %s only play %d matches. This mean you have to relax the requirements to generate enough matching", player, matchesPlayed)
			return nil, err
		}
	}

	// TODO - Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
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

func generateKey(player1, player2 string) string {
	if player1 > player2 {
		return player1 + "-" + player2
	}
	return player2 + "-" + player1
}

func isPlayerExistInList(playerList []string, match SoloHunterMatch) bool {
	return slices.Contains(playerList, match.Player1) ||
		slices.Contains(playerList, match.Player2) ||
		slices.Contains(playerList, match.Player3) ||
		slices.Contains(playerList, match.Player4)
}
