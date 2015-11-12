package main

import (
	"encoding/binary"
	"flag"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gdraynz/go-discord/discord"
)

var (
	reminderFlagDB = flag.String("reminderdb", "reminder.db", "DB file for reminders")
)

// "UserID": {
//		"reminder": {
//			"StartTime": int,
//			"RemindAt": int,
//			"Message": string,
//			"UserID": string,
//		}
// }

type Reminder struct {
	DB        *bolt.DB
	UserID    string
	StartTime time.Time
	RemindAt  time.Time
	Message   string
}

func NewReminder(DB *bolt.DB, userID string, remindIn time.Duration, message string) error {

	// Add reminder to DB (key, no value)
	if err := reminder.DB.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists(reminder.UserID)
		if err != nil {
			return err
		}
		if err := b.Put(reminder.RemindKey(), []byte(reminder.Message)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	log.Printf("New reminder for %s", userID)
	reminder.Start()
	return nil
}

func (reminder *Reminder) RemindKey() []byte {
	remindKey := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(remindKey, reminder.RemindAt.UnixNano())
	return remindKey
}

func (reminder *Reminder) Start() {
	time.AfterFunc(reminder.RemindAt.Sub(time.Now()), func() {
		if err := reminder.Stop(); err != nil {
			log.Printf("error Stop Reminder: %s", err)
		}
	})
}

func (reminder *Reminder) Stop() error {
	// client.GetPri
	return reminder.DB.Update(func(t *bolt.Tx) (err error) {
		b := t.Bucket([]byte(reminder.UserID))
		err = b.Delete(reminder.RemindKey())
		return err
	})
}

type TimeReminder struct {
	DB *bolt.DB
}

func NewTimeReminder() (*TimeReminder, error) {
	var tr *TimeReminder

	db, err := bolt.Open(*gametimeFlagDB, 0600, nil)
	if err != nil {
		return tr, err
	}

	return &TimeReminder{
		DB: db,
	}, nil
}

func (tr *TimeReminder) NewReminderFromBucket(bucket *bolt.Bucket) {
	remindAt, _ := binary.Varint(bucket.Get([]byte("RemindAt")))
	if remindAt < time.Now().Unix() {
		log.Print("Reminder already gone")
		return
	}

	startTime, _ := binary.Varint(bucket.Get([]byte("StartTime")))
	message := bucket.Get([]byte("Message"))
	userID := bucket.Get([]byte("UserID"))

	reminder := Reminder{
		DB:        tr.DB,
		UserID:    string(bucket.Get([]byte("UserID"))[:]),
		StartTime: time.Unix(0, startTime),
		RemindAt:  time.Unix(0, remindAt),
		Message:   string(bucket.Get([]byte("Message"))[:]),
	}

	reminder.Start()
}

func (tr *TimeReminder) NewReminder(userID string, remindIn time.Duration, message string) {

}

func (tr *TimeReminder) ReloadDB() error {
	return tr.DB.View(func(t *bolt.Tx) error {
		// Search through each user
		t.ForEach(func(id []byte, b *bolt.Bucket) error {
			userID := string(name[:])
			// Search through each reminder
			b.ForEach(func(reminderBucket []byte, value []byte) error {
				if message == nil {
					rBucket, _ := b.CreateBucketIfNotExists(reminderBucket)
					tr.NewReminderFromBucket(rBucket)
				} else {
					log.Print("not a bucket, should not happen")
				}
			})
			return nil
		})
		return nil
	})
}

func (tr *TimeReminder) Remind(user *discord.User, duration time.Duration, message string) error {
	return NewReminder(tr.DB, user.ID, duration, message)
}
