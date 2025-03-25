package main

import (
	log "github.com/sirupsen/logrus"
)

// Conditions for solo hunter format
const SOLO_HUNTER_TOTAL_MATCH_PER_PERSON = 8
const SOLO_HUNTER_MAX_RANK_PERCENTAGE_DIFFERENCE = 0.15
const SOLO_HUNTER_MAX_REPEATED_OPPONENT = 2
const SOLO_HUNTER_RERUN = 50000

// Conditions for match maker format
const MATCH_MAKER_TOTAL_MATCH_PER_PERSON = 3
const MATCH_MAKER_MAX_RANK_PERCENTAGE_DIFFERENCE = 0.1
const MATCH_MAKER_MAX_REPEATED_OPPONENT = 1
const MATCH_MAKER_RERUN = 10000

const NUMBER_OF_COURT = 4

type PlayerMetadata struct {
	Name        string  `csv:"name"`
	Gender      string  `csv:"gender"`
	TitRanking  float64 `csv:"tit_ranking"`
	TaRanking   float64 `csv:"ta_ranking"`
	MinhRanking float64 `csv:"minh_ranking"`
}

func main() {
	log.SetReportCaller(true)
	_, err := generateSoloHunterMatchesUntilSuccess()
	if err != nil {
		log.Errorf("fail to generate solo hunter matches %v", err)
	}

	_, err = generateMatchMakerMatchesUntilSuccess()
	if err != nil {
		log.Errorf("failed to generate match maker matches %v", err)
	}
}
