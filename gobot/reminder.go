package main

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gdraynz/go-discord/discord"
	"github.com/satori/go.uuid"
)

var (
	reminderFlagDB = flag.String("reminderdb", "reminder.db", "DB file for reminders")
)

// "<user id>": {
//		"<uuid>": {
//			"RemindAt": "<unix time in seconds>",
//			"Message": "<message>",
//			"UserID": "<user id>",
//		}
// }

type Reminder struct {
	DB       *bolt.DB
	UUID     string
	UserID   string
	RemindAt time.Time
	Message  string
}

func (reminder *Reminder) Save() error {
	return reminder.DB.Update(func(t *bolt.Tx) (err error) {
		userBucket, err := t.CreateBucketIfNotExists([]byte(reminder.UserID))
		if err != nil {
			return err
		}
		rBucket, _ := userBucket.CreateBucketIfNotExists([]byte(reminder.UUID))

		bRemind, _ := reminder.RemindAt.MarshalBinary()
		err = rBucket.Put([]byte("RemindAt"), bRemind)
		err = rBucket.Put([]byte("Message"), []byte(reminder.Message))
		err = rBucket.Put([]byte("UserID"), []byte(reminder.UserID))

		return err
	})
}

func (reminder *Reminder) Start() {
	time.AfterFunc(reminder.RemindAt.Sub(time.Now()), func() {
		if err := reminder.Ping(); err != nil {
			log.Printf("error Stop Reminder: %s", err)
		}
	})
}

func (reminder *Reminder) Ping() error {
	// Send private message
	pc := client.GetPrivateChannel(client.GetUserByID(reminder.UserID))
	pc.SendMessage(&client, reminder.Message)

	log.Print("Reminder sent, deleting entry")

	// Delete DB entry
	return reminder.DB.Update(func(t *bolt.Tx) (err error) {
		b := t.Bucket([]byte(reminder.UserID))
		return b.DeleteBucket([]byte(reminder.UUID))
	})

}

type TimeReminder struct {
	RemindersSent int
	DB            *bolt.DB
}

func NewTimeReminder() (*TimeReminder, error) {
	var tr *TimeReminder

	db, err := bolt.Open(*reminderFlagDB, 0600, nil)
	if err != nil {
		return tr, err
	}

	tr = &TimeReminder{
		DB: db,
	}

	if err := tr.ReloadDB(); err != nil {
		return tr, err
	}

	return tr, nil
}

func (tr *TimeReminder) NewReminder(user discord.User, remindIn time.Duration, message string) {
	reminder := Reminder{
		UUID:     uuid.NewV4().String(),
		DB:       tr.DB,
		UserID:   user.ID,
		RemindAt: time.Now().Add(remindIn),
		Message:  message,
	}

	if err := reminder.Save(); err != nil {
		log.Print(err)
	} else {
		reminder.Start()
	}
}

func (tr *TimeReminder) NewReminderFromBucket(bUID []byte, bucket *bolt.Bucket) error {
	var remindAt time.Time

	userID := string(bucket.Get([]byte("UserID"))[:])

	remindAt.UnmarshalBinary(bucket.Get([]byte("RemindAt")))
	if remindAt.Unix() < time.Now().Unix() {
		log.Print("Reminder already gone")
		return errors.New("reminder gone")
	}

	reminder := Reminder{
		UUID:     string(bUID),
		DB:       tr.DB,
		UserID:   userID,
		RemindAt: remindAt,
		Message:  string(bucket.Get([]byte("Message"))[:]),
	}

	reminder.Start()
	return nil
}

func (tr *TimeReminder) GetUserReminders(user discord.User) (map[string]time.Time, error) {
	// res := make(map[string])
	err := tr.DB.View(func(t *bolt.Tx) error {

	})
	if err != nil {
		return nil
	}
}

func (tr *TimeReminder) ReloadDB() error {
	return tr.DB.Update(func(t *bolt.Tx) error {
		// Search through each user
		t.ForEach(func(userID []byte, b *bolt.Bucket) error {
			// Search through each reminder
			b.ForEach(func(bUID []byte, value []byte) error {
				if value == nil {
					rBucket := b.Bucket(bUID)
					if rBucket != nil {
						if err := tr.NewReminderFromBucket(bUID, rBucket); err != nil {
							b.DeleteBucket(bUID)
						}
					}
				} else {
					log.Print("not a bucket, should not happen")
				}
				return nil
			})
			return nil
		})
		return nil
	})
}
