package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

// Conditions for match maker format
const MATCH_MAKER_TOTAL_MATCH_PER_PERSON = 8
const MATCH_MAKER_MAX_RANK_DIFFERENCE = 0.6
const MATCH_MAKER_MAX_REPEATED_OPPONENT = 2
const MATCH_MAKER_RERUN = 10000

// const MATCH_MAKER_MAX_TEAM_TOGETHER = 2

type PlayerRanking struct {
	Name        string  `csv:"name"`
	TitRanking  float64 `csv:"tit_ranking"`
	TaRanking   float64 `csv:"ta_ranking"`
	MinhRanking float64 `csv:"minh_ranking"`
}

type MatchMakerMatch struct {
	player1 string `csv:"player1"`
	player2 string `csv:"player2"`
	player3 string `csv:"player3"`
	player4 string `csv:"player4"`
}

func main() {
	log.SetReportCaller(true)
	_, err := generateMatchMakerMatchesUntilSuccess()
	if err != nil {
		panic(err)
	}
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
		log.Infof("%s,%s,%s,%s", match.player1, match.player2, match.player3, match.player4)
	}
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

	matchMakerMatches := []MatchMakerMatch{}
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
					if isAllPlayersDifferent([]string{player1, player2, player3, player4}) &&
						math.Abs(playerAndRanking[player1]+playerAndRanking[player2]-playerAndRanking[player3]-playerAndRanking[player4]) < MATCH_MAKER_MAX_RANK_DIFFERENCE &&
						matchMakerPlayerAndNoOfMatch[player1] <= MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player2] <= MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player3] <= MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndNoOfMatch[player4] <= MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player3)] <= MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player1, player4)] <= MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player3)] <= MATCH_MAKER_MAX_REPEATED_OPPONENT &&
						matchMakerPlayerAndOponentNoOfMatch[generateKey(player2, player4)] <= MATCH_MAKER_MAX_REPEATED_OPPONENT {
						matchMakerMatches = append(matchMakerMatches, MatchMakerMatch{
							player1: player1,
							player2: player2,
							player3: player3,
							player4: player4,
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
	// TODO - Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
	matchMakerFile, err := os.OpenFile(fmt.Sprintf("matchMaker_%s.csv", time.Now().Format("20060102150405")), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer matchMakerFile.Close()
	if err = gocsv.MarshalFile(&matchMakerMatches, matchMakerFile); err != nil {
		panic(err)
	}

	// Check if player play enough match
	for player, matchesPlayed := range matchMakerPlayerAndNoOfMatch {
		if matchesPlayed < MATCH_MAKER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("player %s only play %d matches. This mean you have to relax the requirements to generate enough matching", player, matchesPlayed)
			return nil, err
		}
	}

	return matchMakerMatches, nil
}

func loadAvgRanking() (map[string]float64, error) {
	// Load from csv and take average of all ranking score. also output average team ranking
	rankingFile, err := os.OpenFile("ranking_score.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rankingFile.Close()

	playerRankings := []*PlayerRanking{}
	if err := gocsv.UnmarshalFile(rankingFile, &playerRankings); err != nil {
		// Load clients from file
		log.Error(err)
		return nil, err
	}

	playerAvgRanking := make(map[string]float64)
	for _, client := range playerRankings {
		playerAvgRanking[client.Name] = (client.TitRanking + client.TaRanking + client.MinhRanking) / 3
	}
	return playerAvgRanking, nil
}

func generateKey(player1, player2 string) string {
	if player1 > player2 {
		return player1 + "-" + player2
	}
	return player2 + "-" + player1
}

func isAllPlayersDifferent(players []string) bool {
	playerIsUnique := make(map[string]bool)
	for _, player := range players {
		if _, ok := playerIsUnique[player]; ok {
			return false
		}
		playerIsUnique[player] = true
	}
	return true
}
