package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gdraynz/go-discord/discord"
)

var (
	flagPlayed = flag.String("played", "played.json", "Played time dump json file")
	flagDB     = flag.String("db", "played.db", "DB file for game time")

	playTimeDB *bolt.DB
)

type TimeCounter struct {
	InProgress map[string]chan bool
}

func NewCounter() (_ *TimeCounter, err error) {
	playTimeDB, err = bolt.Open(*flagDB, 0600, nil)
	return &TimeCounter{InProgress: make(map[string]chan bool)}, err
}

func (t *TimeCounter) GetUserGameTime(user discord.User) {

}

func (t *TimeCounter) CountGameTime(user discord.User, game discord.Game) {
	log.Printf("Starting to count for %s on %s", user.Name, game.Name)

	gameid := string(game.ID)

	t.InProgress[user.ID] = make(chan bool)
	start := time.Now()

	// Wait for game to end
	<-t.InProgress[user.ID]

	// Delete user from playing list
	delete(t.InProgress, user.ID)

	// Update game time
	playTimeDB.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists([]byte(user.ID))
		if err != nil {
			return err
		}
		bPlayed := b.Get([]byte(gameid))
		if bPlayed != nil {
			// bytes to int64
			played, _ := binary.Varint(bPlayed)
			// Calc total time
			total := time.Now().Add(time.Duration(played))
			var newPlayed []byte
			// int64 to bytes
			binary.PutVarint(&newPlayed, total.Sub(start).Nanoseconds())
			b.Put([]byte(gameid), newPlayed)
		} else {
			var newPlayed []byte
			binary.PutVarint(&newPlayed, time.Since(start).Nanoseconds())
			b.Put([]byte(gameid), newPlayed)
		}
		return nil
	})

	log.Printf("Done counting for %s", user.Name)
}

func (t *TimeCounter) LoadGameTime() error {
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

func (t *TimeCounter) SaveGameTime() {
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
