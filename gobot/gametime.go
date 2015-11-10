package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gdraynz/go-discord/discord"
)

var (
	flagDB = flag.String("db", "gametime.db", "DB file for game time")

	GametimeDB *bolt.DB
)

type TimeCounter struct {
	InProgress map[string]chan bool
}

func NewCounter() (_ *TimeCounter, err error) {
	GametimeDB, err = bolt.Open(*flagDB, 0600, nil)
	return &TimeCounter{InProgress: make(map[string]chan bool)}, err
}

func (t *TimeCounter) GetUserGametime(user discord.User) (map[string]int64, error) {
	gameMap := make(map[string]int64)
	err := GametimeDB.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(user.ID))
		if b == nil {
			return errors.New("user never played")
		}
		// Iterate through all games
		b.ForEach(func(gameID []byte, nanoTime []byte) error {
			gameMap[string(gameID[:])], _ = binary.Varint(nanoTime)
			return nil
		})
		return nil
	})
	return gameMap, err
}

func (t *TimeCounter) CountGametime(user discord.User, game discord.Game) {
	log.Printf("Starting to count for %s on %s", user.Name, game.Name)

	gameid := string(game.ID)

	t.InProgress[user.ID] = make(chan bool)
	start := time.Now()

	// Wait for game to end
	<-t.InProgress[user.ID]

	// Delete user from playing list
	delete(t.InProgress, user.ID)

	// Update game time
	err := GametimeDB.Update(func(t *bolt.Tx) error {
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

			// int64 to bytes
			newPlayed := make([]byte, binary.MaxVarintLen64)
			binary.PutVarint(newPlayed, total.Sub(start).Nanoseconds())

			if err := b.Put([]byte(gameid), newPlayed); err != nil {
				return err
			}
		} else {
			// int64 to bytes
			newPlayed := make([]byte, binary.MaxVarintLen64)
			binary.PutVarint(newPlayed, time.Since(start).Nanoseconds())

			if err := b.Put([]byte(gameid), newPlayed); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error while updating game time : %s", err.Error())
	} else {
		log.Printf("Done counting for %s", user.Name)
	}
}

func (t *TimeCounter) Close() {
	GametimeDB.Close()
}
