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
	teams, err := teamAssign()
	if err != nil {
		return nil, err
	}

	runNo := 0
	matches, err := generateMatchMakerMatches(teams)
	for err != nil && runNo < MATCH_MAKER_RERUN {
		log.Infof("Running for round %d", runNo)
		matches, err = generateMatchMakerMatches(teams)
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

func teamAssign() ([]MatchMakerTeam, error) {
	// Now divided players into teams so that best pair with worst
	teams := MATCH_MAKER_PREFERENCE_LIST
	playersToAssignSorted := sortPlayersByRanking()
	player1Pointer := 0
	player2Pointer := len(playersToAssignSorted)-1
	for player1Pointer < player2Pointer {
		player1 := playersToAssignSorted[player1Pointer]
		player2 := playersToAssignSorted[player2Pointer]
		if isPlayerAlreadyHasPreference(player1.Name) {
			player1Pointer++
			continue
		}
		if isPlayerAlreadyHasPreference(player2.Name) {
			player2Pointer--
			continue
		}

		if player1.Gender == "f" && player2.Gender == "f" {
			if player1Pointer + 1 == player2Pointer {
				err := fmt.Errorf("impossible case, since there is 2 female players at the middle, need to adjust PREFERENCE LIST to avoid this, list of players sorted by ranking %v", playersToAssignSorted)
				log.Error(err)
				return nil, err
			}
			// TODO what if there is no male left, too many female??
			for i := player2Pointer-1; i>player1Pointer; i-- {
				if playersToAssignSorted[i].Gender == "m" {
					playersToAssignSorted[i], playersToAssignSorted[player2Pointer] = playersToAssignSorted[player2Pointer], playersToAssignSorted[i]
					player2 = playersToAssignSorted[player2Pointer]
				}
			}
		}		
		teams = append(teams, MatchMakerTeam{
			Player1: playersToAssignSorted[player1Pointer].Name,
			Player2: playersToAssignSorted[player2Pointer].Name,
		})
		player1Pointer++;
		player2Pointer--;
	}
	log.Infof("Teams = %v", teams)


	return teams, nil
}

func generateMatchMakerMatches(teams []MatchMakerTeam) ([]MatchMetadata, error) {
	playerAndRanking, _, err := loadAvgRankingAndGender()
	if err != nil {
		log.Fatal(err)
	}

	matches := []MatchMetadata{}
	playerAndNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		playerAndNoOfMatch[player] = 0
	}
	playerAndOponentNoOfMatch := make(map[string]int16)
	for player := range playerAndRanking {
		for opponent := range playerAndRanking {
			if opponent != player {
				playerAndOponentNoOfMatch[generateKey(player, opponent)] = 0
			}
		}
	}
	
	for _, team1 := range teams {
		// Randomly shuffle the teams order
		teamRandomizeOrder := make([]MatchMakerTeam, len(teams))
		copy(teamRandomizeOrder, teams)
		for i := range teamRandomizeOrder {
			j := rand.Intn(i + 1)
			teamRandomizeOrder[i], teamRandomizeOrder[j] = teamRandomizeOrder[j], teamRandomizeOrder[i]
		}
		for _, team2 := range teamRandomizeOrder {
			opponentKey := generateKey(team1.Player1, team2.Player1)
			if team1.Player1 != team2.Player1 &&
				playerAndNoOfMatch[team1.Player1] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
				playerAndNoOfMatch[team2.Player1] < MATCH_MAKER_TOTAL_MATCH_PER_PERSON &&
				playerAndOponentNoOfMatch[opponentKey] < MATCH_MAKER_MAX_REPEATED_OPPONENT {
				matches = append(matches, MatchMetadata{
					Player1: team1.Player1,
					Player2: team1.Player2,
					Player3: team2.Player1,
					Player4: team2.Player2,
					PercentageDifference: percentageDifference(team1.Player1, team1.Player2, team2.Player1, team2.Player2, playerAndRanking),
				})
				playerAndNoOfMatch[team1.Player1] = playerAndNoOfMatch[team1.Player1] + 1
				playerAndNoOfMatch[team2.Player1] = playerAndNoOfMatch[team2.Player1] + 1
				playerAndOponentNoOfMatch[opponentKey] = playerAndOponentNoOfMatch[opponentKey] + 1
			}
		}
	}

	// Check if teams play enough match
	for _, team := range teams {
		if playerAndNoOfMatch[team.Player1] != MATCH_MAKER_TOTAL_MATCH_PER_PERSON {
			err = fmt.Errorf("team %v plays %d matches. This mean you have to relax the requirements to generate enough matching, or check why player has to play more than they should", team, playerAndNoOfMatch[team.Player1])
			log.Error(err)
			return nil, err
		}
	}

	// Assign match to court so that no player have to player 2 match in one round
	matches, err = assignMatchesToCourts(matches)
	if err != nil {
		return nil, err
	}

	// Remember to output the result into a csv file, with timestamp to version control and allow us to the best possible match-up
	MatchMakerFile, err := os.OpenFile(fmt.Sprintf("MatchMaker_%s.csv", time.Now().Format("200601021504")), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer MatchMakerFile.Close()
	if err = gocsv.MarshalFile(&matches, MatchMakerFile); err != nil {
		panic(err)
	}

	return matches, nil
}

func isPlayerAlreadyHasPreference(player string) bool {
	for _, preferredTeam := range MATCH_MAKER_PREFERENCE_LIST {
		if player == preferredTeam.Player1 || player == preferredTeam.Player2 {
			return true
		}
	}
	return false
}
