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
	gametimeFlagDB = flag.String("gametimedb", "gametime.db", "DB file for game time")
)

// TODO: migrate that in seconds...
// "<user id>": {
//		"<game id>": "<time in nanoseconds>"
// }

type PlayingUser struct {
	UserID    string
	StartTime time.Time
	GameID    string
}

func (p *PlayingUser) SaveGametime(t *bolt.Tx) error {
	b, err := t.CreateBucketIfNotExists([]byte(p.UserID))
	if err != nil {
		return err
	}
	bPlayed := b.Get([]byte(p.GameID))
	if bPlayed != nil {
		// bytes to int64
		played, _ := binary.Varint(bPlayed)

		// Calc total time
		total := time.Now().Add(time.Duration(played))

		// int64 to bytes
		newPlayed := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(newPlayed, total.Sub(p.StartTime).Nanoseconds())

		if err := b.Put([]byte(p.GameID), newPlayed); err != nil {
			return err
		}
	} else {
		// int64 to bytes
		newPlayed := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(newPlayed, time.Since(p.StartTime).Nanoseconds())

		if err := b.Put([]byte(p.GameID), newPlayed); err != nil {
			return err
		}
	}
	// Update start time
	p.StartTime = time.Now()
	return nil
}

type GametimeCounter struct {
	InProgress   map[string]PlayingUser
	DB           *bolt.DB
	GametimeChan chan PlayingUser
}

func NewCounter() (*GametimeCounter, error) {
	var t *GametimeCounter

	db, err := bolt.Open(*gametimeFlagDB, 0600, nil)
	if err != nil {
		return t, err
	}

	return &GametimeCounter{
		InProgress:   make(map[string]PlayingUser),
		DB:           db,
		GametimeChan: make(chan PlayingUser),
	}, nil
}

func (counter *GametimeCounter) Listen() {
	for {
		pUser := <-counter.GametimeChan
		go counter.EndGametime(pUser)
	}
}

func (counter *GametimeCounter) StartGametime(user discord.User, game discord.Game) {
	log.Printf("Starting to count for %s on %s", user.Name, game.Name)

	pUser := PlayingUser{
		UserID:    user.ID,
		GameID:    string(game.ID),
		StartTime: time.Now(),
	}

	counter.InProgress[user.ID] = pUser
}

func (counter *GametimeCounter) EndGametime(pUser PlayingUser) {
	// Delete user from playing list
	delete(counter.InProgress, pUser.UserID)

	// Update game time
	err := counter.DB.Update(func(t *bolt.Tx) error {
		err := pUser.SaveGametime(t)
		return err
	})

	if err != nil {
		log.Printf("Error while updating game time : %s", err.Error())
	}

	log.Printf("Saved %s", pUser.UserID)
}

func (counter *GametimeCounter) ServerGametime(server discord.Server) (map[string]int64, error) {
	gameMap := make(map[string]int64)
	err := counter.DB.View(func(t *bolt.Tx) error {
		for _, member := range server.Members {
			b := t.Bucket([]byte(member.User.ID))
			if b == nil {
				continue
			}
			// Iterate through all games
			b.ForEach(func(gameID []byte, nanoTime []byte) error {
				_, ok := gameMap[string(gameID[:])]
				if ok {
					add, _ := binary.Varint(nanoTime)
					gameMap[string(gameID[:])] += add
				} else {
					gameMap[string(gameID[:])], _ = binary.Varint(nanoTime)
				}
				return nil
			})
		}
		return nil
	})
	return gameMap, err
}

func (counter *GametimeCounter) ResetGametime(user discord.User) error {
	return counter.DB.Update(func(t *bolt.Tx) error {
		return t.DeleteBucket([]byte(user.ID))
	})
}

func (counter *GametimeCounter) ResetOneGametime(user discord.User, game discord.Game) error {
	return counter.DB.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(user.ID))
		if b == nil {
			return errors.New("User unknown")
		}
		// If only one game, delete bucket
		if b.Stats().KeyN == 1 {
			return t.DeleteBucket([]byte(user.ID))
		}
		return b.Delete([]byte(string(game.ID)))
	})
}

func (counter *GametimeCounter) GetUserGametime(user discord.User) (map[string]int64, error) {
	gameMap := make(map[string]int64)
	err := counter.DB.View(func(t *bolt.Tx) error {
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

func (counter *GametimeCounter) Snapshot() error {
	return counter.DB.Update(func(t *bolt.Tx) (err error) {
		for _, pUser := range counter.InProgress {
			err = pUser.SaveGametime(t)
			if err != nil {
				log.Print(err)
				continue
			}
		}
		log.Print("Snapshot done")
		return nil
	})
}

func (counter *GametimeCounter) Close() {
	// Save times for currently playing users
	if err := counter.Snapshot(); err != nil {
		log.Print(err)
	}
	counter.DB.Close()
}
