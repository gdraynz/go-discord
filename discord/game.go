package discord

import (
	"encoding/json"
	"io/ioutil"
)

type Game struct {
	Name        string              `json:"name"`
	ID          int                 `json:"id"`
	Executables map[string][]string `json:"executables"`
}

func GetGamesFromFile(filename string) (map[int]Game, error) {
	fileDump, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var games []Game
	if err := json.Unmarshal(fileDump, &games); err != nil {
		return nil, err
	}

	gameMap := make(map[int]Game)
	for _, game := range games {
		gameMap[game.ID] = game
	}

	return gameMap, nil
}
