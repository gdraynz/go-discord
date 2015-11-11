package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gdraynz/go-discord/discord"
)

var (
	flagDB = flag.String("db", "gametime.db", "DB file for game time")
)

type sortedMap struct {
	m map[string]int64
	s []string
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.s[i]] > sm.m[sm.s[j]]
}

func (sm *sortedMap) Swap(i, j int) {
	sm.s[i], sm.s[j] = sm.s[j], sm.s[i]
}

func sortedKeys(m map[string]int64) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key, _ := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

type PlayingUser struct {
	UserID    string
	StartTime time.Time
	C         chan bool
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

type TimeCounter struct {
	InProgress map[string]PlayingUser
	GametimeDB *bolt.DB
}

func NewCounter() (*TimeCounter, error) {
	var t *TimeCounter

	db, err := bolt.Open(*flagDB, 0600, nil)
	if err != nil {
		return t, err
	}

	return &TimeCounter{
		InProgress: make(map[string]PlayingUser),
		GametimeDB: db,
	}, nil
}

func (counter *TimeCounter) GetUserGametime(user discord.User) (map[string]int64, error) {
	gameMap := make(map[string]int64)
	err := counter.GametimeDB.View(func(t *bolt.Tx) error {
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

type TopGame struct {
	ID         string
	TimePlayed int64
}

func (counter *TimeCounter) GetTopGames() ([]TopGame, error) {
	gameMap := make(map[string]int64)

	err := counter.GametimeDB.View(func(t *bolt.Tx) error {
		t.ForEach(func(bName []byte, b *bolt.Bucket) error {
			b.ForEach(func(gameID []byte, gameTime []byte) error {
				strGameID := string(gameID[:])
				addTime, _ := binary.Varint(gameTime)

				time, ok := gameMap[strGameID]
				if ok {
					time += addTime
				} else {
					gameMap[strGameID] = addTime
				}

				return nil
			})
			return nil
		})
		return nil
	})

	sorted := sortedKeys(gameMap)
	if len(sorted)-1 < 3 {
		return nil, errors.New("failed to sort")
	}

	var top []TopGame
	for i := 0; i < 3; i++ {
		top = append(top, TopGame{
			ID:         sorted[i],
			TimePlayed: gameMap[sorted[i]],
		})
	}

	return top, err
}

func (counter *TimeCounter) CountGametime(user discord.User, game discord.Game) {
	log.Printf("Starting to count for %s on %s", user.Name, game.Name)

	pUser := PlayingUser{
		UserID:    user.ID,
		GameID:    string(game.ID),
		C:         make(chan bool),
		StartTime: time.Now(),
	}
	counter.InProgress[user.ID] = pUser

	// Wait for game to end
	<-counter.InProgress[user.ID].C

	// Delete user from playing list
	delete(counter.InProgress, user.ID)

	// Update game time
	err := counter.GametimeDB.Update(func(t *bolt.Tx) error {
		err := pUser.SaveGametime(t)
		log.Printf("%s saved", user.Name)
		return err
	})

	if err != nil {
		log.Printf("Error while updating game time : %s", err.Error())
	}
}

func (counter *TimeCounter) Snapshot() error {
	return counter.GametimeDB.Update(func(t *bolt.Tx) (err error) {
		for _, pUser := range counter.InProgress {
			err = pUser.SaveGametime(t)
			if err != nil {
				log.Print(err)
				continue
			}
		}
		return nil
	})
}

func (counter *TimeCounter) Close() {
	// Save times for currently playing users
	if err := counter.Snapshot(); err != nil {
		log.Print(err)
	}
	counter.GametimeDB.Close()
}
