package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"slices"
	"sort"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

type MatchMetadata struct {
	MatchNo int `csv:"match_no"`
	Court                int     `csv:"court"`
	Player1              string  `csv:"player1"`
	Player2              string  `csv:"player2"`
	Player3              string  `csv:"player3"`
	Player4              string  `csv:"player4"`
	PercentageDifference float64 `csv:"percentage_difference"`
}

type MatchMakerTeam struct {
	Player1 string
	Player2 string
}

type PlayerMetadata struct {
	Name        string  `csv:"name"`
	Gender      string  `csv:"gender"`
	TitRanking  float64 `csv:"tit_ranking"`
	TaRanking   float64 `csv:"ta_ranking"`
	MinhRanking float64 `csv:"minh_ranking"`
}

func loadAvgRankingAndGender() (map[string]float64, map[string]string, error) {
	// Load from csv and take average of all ranking score. also output average team ranking
	rankingFile, err := os.OpenFile("ranking_score.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	defer rankingFile.Close()

	playersMetadatas := []*PlayerMetadata{}
	if err := gocsv.UnmarshalFile(rankingFile, &playersMetadatas); err != nil {
		// Load clients from file
		log.Error(err)
		return nil, nil, err
	}

	playerAvgRanking := make(map[string]float64)
	for _, client := range playersMetadatas {
		playerAvgRanking[client.Name] = (client.TitRanking + client.TaRanking + client.MinhRanking) / 3
	}

	playerGender := make(map[string]string)
	for _, player := range playersMetadatas {
		playerGender[player.Name] = player.Gender
	}
	return playerAvgRanking, playerGender, nil
}

func sortPlayersByRanking() []PlayerMetadata {
	// Load from csv and take average of all ranking score. also output average team ranking
	rankingFile, err := os.OpenFile("ranking_score.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer rankingFile.Close()

	playersMetadatas := []PlayerMetadata{}
	if err := gocsv.UnmarshalFile(rankingFile, &playersMetadatas); err != nil {
		// Load clients from file
		panic(err)
	}

	sort.Slice(playersMetadatas, func(i, j int) bool {
		return playersMetadatas[i].MinhRanking + playersMetadatas[i].TaRanking + playersMetadatas[i].TitRanking > playersMetadatas[j].MinhRanking + playersMetadatas[j].TaRanking + playersMetadatas[j].TitRanking
	})
	return playersMetadatas
}

func isAllPlayersDifferentAndNoTwoFemaleSameTeam(players []string, playerGender map[string]string) bool {
	playerIsUnique := make(map[string]bool)
	for _, player := range players {
		if _, ok := playerIsUnique[player]; ok {
			return false
		}
		playerIsUnique[player] = true
	}

	if (playerGender[players[0]] == "f" && playerGender[players[1]] == "f") || (playerGender[players[2]] == "f" && playerGender[players[3]] == "f") {
		return false
	}
	return true
}

func generateKey(player1, player2 string) string {
	if player1 > player2 {
		return player1 + "-" + player2
	}
	return player2 + "-" + player1
}

func isPlayerExistInList(playerList []string, match MatchMetadata) bool {
	return slices.Contains(playerList, match.Player1) ||
		slices.Contains(playerList, match.Player2) ||
		slices.Contains(playerList, match.Player3) ||
		slices.Contains(playerList, match.Player4)
}

func percentageDifference(player1, player2, player3, player4 string, playerAndRanking map[string]float64) float64 {
	rankDistance := math.Abs(playerAndRanking[player1] + playerAndRanking[player2] - playerAndRanking[player3] - playerAndRanking[player4])
	rankOfLesserTeam := math.Min(playerAndRanking[player1]+playerAndRanking[player2], playerAndRanking[player3]+playerAndRanking[player4])

	return rankDistance / rankOfLesserTeam
}

func assignMatchesToCourts(matchMakerMatches []MatchMetadata, listOfCourts []int, disableMatchShuffle bool) ([]MatchMetadata, error) {
	// Randomly shuffle the matches order
	if !disableMatchShuffle {
		for i := range matchMakerMatches {
			j := rand.Intn(i + 1)
			matchMakerMatches[i], matchMakerMatches[j] = matchMakerMatches[j], matchMakerMatches[i]
		}
	}

	playersInCurrentRound := []string{}
	// TODO - Change this later, for simplicity, assume there is only 4 courts right now
	indexInListOfCourts := 0
	matchNo := 1
	for i := range matchMakerMatches {
		if indexInListOfCourts == 0 {
			playersInCurrentRound = []string{}
		}

		// If any of 4 player already play this round, find closest group of 4 players that hasn't play and swap the index
		if isPlayerExistInList(playersInCurrentRound, matchMakerMatches[i]) {
			foundMatch := false
			for j := i + 1; j < len(matchMakerMatches); j++ {
				if !isPlayerExistInList(playersInCurrentRound, matchMakerMatches[j]) {
					matchMakerMatches[i], matchMakerMatches[j] = matchMakerMatches[j], matchMakerMatches[i]
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				// Skip this round and go straight to next round - 
				// indexInListOfCourts = 0
				// matchMakerMatches[i].Court = listOfCourts[indexInListOfCourts]
				// matchMakerMatches[i].MatchNo = matchNo
				// playersInCurrentRound = []string{matchMakerMatches[i].Player1, matchMakerMatches[i].Player2, matchMakerMatches[i].Player3, matchMakerMatches[i].Player4}
				// indexInListOfCourts = (indexInListOfCourts + 1) % len(listOfCourts)
				// matchNo++
				// continue
				
				err := fmt.Errorf("cannot allocate court for some reason, playersInCurrentRound = %v, court %d remaining matches = %v", playersInCurrentRound, indexInListOfCourts, matchMakerMatches[i+1:])
				log.Error(err)
				return nil, err
			}
		}

		matchMakerMatches[i].Court = listOfCourts[indexInListOfCourts]
		matchMakerMatches[i].MatchNo = matchNo
		playersInCurrentRound = append(playersInCurrentRound, matchMakerMatches[i].Player1, matchMakerMatches[i].Player2, matchMakerMatches[i].Player3, matchMakerMatches[i].Player4)
		indexInListOfCourts = (indexInListOfCourts + 1) % len(listOfCourts)
		matchNo++
	}

	return matchMakerMatches, nil
}