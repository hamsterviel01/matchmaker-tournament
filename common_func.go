package main

import (
	"math"
	"os"
	"slices"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)



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
		return nil, nil, err
	}

	playerAvgRanking := make(map[string]float64)
	for _, client := range playerRankings {
		playerAvgRanking[client.Name] = (client.TitRanking + client.TaRanking + client.MinhRanking) / 3
	}
	return playerAvgRanking, nil, nil
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

func percentageDifference(player1, player2, player3, player4 string, playerAndRanking map[string]float64) float64 {
	rankDistance := math.Abs(playerAndRanking[player1] + playerAndRanking[player2] - playerAndRanking[player3] - playerAndRanking[player4])
	rankOfLesserTeam := math.Min(playerAndRanking[player1] + playerAndRanking[player2], playerAndRanking[player3] + playerAndRanking[player4])

	return rankDistance/rankOfLesserTeam
}