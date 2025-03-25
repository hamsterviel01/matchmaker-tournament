package main

import (
	"os"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

// Conditions for solo hunter format
const SOLO_HUNTER_TOTAL_MATCH_PER_PERSON = 8
const SOLO_HUNTER_MAX_RANK_DIFFERENCE = 0.6
const SOLO_HUNTER_MAX_REPEATED_OPPONENT = 2
const SOLO_HUNTER_RERUN = 10000

// Conditions for match maker format
const MATCH_MAKER_TOTAL_MATCH_PER_PERSON = 3
const MATCH_MAKER_MAX_RANK_DIFFERENCE = 0.6
const MATCH_MAKER_MAX_REPEATED_OPPONENT = 1
const MATCH_MAKER_RERUN = 10000

const NUMBER_OF_COURT = 4

type PlayerRanking struct {
	Name        string  `csv:"name"`
	TitRanking  float64 `csv:"tit_ranking"`
	TaRanking   float64 `csv:"ta_ranking"`
	MinhRanking float64 `csv:"minh_ranking"`
}

func main() {
	log.SetReportCaller(true)
	_, err := generateSoloHunterMatchesUntilSuccess()
	if err != nil {
		log.Errorf("fail to generate solo hunter matches %v",  err)
	}

	_, err = generateMatchMakerMatchesUntilSuccess()
	if err != nil {
		log.Errorf("failed to generate match maker matches %v", err)
	}
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
