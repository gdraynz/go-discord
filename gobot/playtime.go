package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gdraynz/go-discord/discord"
)

var (
	flagPlayed = flag.String("played", "played.json", "Played time dump json file")
)

type TimeCounter struct {
	InProgress map[string]chan bool
	Played     map[string]map[string]int64
}

func NewCounter() *TimeCounter {
	return &TimeCounter{
		InProgress: make(map[string]chan bool),
		Played:     make(map[string]map[string]int64),
	}
}

func (t *TimeCounter) CountPlaytime(user discord.User, game discord.Game) {
	log.Printf("Starting to count for %s on %s", user.Name, game.Name)

	t.InProgress[user.ID] = make(chan bool)
	start := time.Now()
	_, ok := t.Played[user.ID]
	if !ok {
		t.Played[user.ID] = make(map[string]int64)
	}

	// Wait for game to end
	<-t.InProgress[user.ID]

	// Delete user from playing list
	delete(t.InProgress, user.ID)

	// Update player's game time
	gameid := string(game.ID)
	_, alreadyPlayed := t.Played[user.ID][gameid]
	if alreadyPlayed {
		total := time.Now().Add(time.Duration(t.Played[user.ID][gameid]))
		t.Played[user.ID][gameid] = total.Sub(start).Nanoseconds()
	} else {
		t.Played[user.ID][gameid] = time.Since(start).Nanoseconds()
	}

	log.Printf("Done counting for %s", user.Name)
}

func (t *TimeCounter) LoadPlayTime() error {
	_, err := os.Stat(*flagPlayed)

	if err != nil {
		_, err := os.OpenFile(*flagPlayed, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
	} else {
		dump, err := ioutil.ReadFile(*flagPlayed)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dump, &t.Played); err != nil {
			return err
		}
	}

	return nil
}

func (t *TimeCounter) SavePlayTime() {
	for _, c := range t.InProgress {
		c <- true
	}

	// Wait 10ms to save all play times (purely speculative)
	time.Sleep(10 * time.Millisecond)

	dump, err := json.Marshal(t.Played)
	if err != nil {
		log.Print(err)
		return
	}
	if err := ioutil.WriteFile(*flagPlayed, dump, 0600); err != nil {
		log.Print(err)
	}
}
