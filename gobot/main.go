package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gdraynz/go-discord/discord"
)

const (
	errorMessage = "Woops. I failed :("
)

var (
	flagConf   = flag.String("conf", "conf.json", "Configuration file")
	flagStdout = flag.Bool("stdout", true, "Logs to stdout")

	client        discord.Client
	startTime     time.Time
	commands      map[string]Command
	games         map[string]discord.Game
	totalCommands int

	counter  *GametimeCounter
	reminder *TimeReminder
)

type Command struct {
	Word    string
	Help    string
	Handler func(discord.Message, ...string)
}

func getGameByName(gameName string) (res discord.Game, err error) {
	found := false
	lowerGameName := strings.ToLower(gameName)
	for _, game := range games {
		if strings.ToLower(game.Name) == lowerGameName {
			res = game
			found = true
		}
	}
	if !found {
		err = errors.New("game not found")
	}
	return res, err
}

func onReady(ready discord.Ready) {
	startTime = time.Now()
	totalCommands = 0

	// Init game list
	var err error
	games, err = discord.GetGamesFromFile("games.json")
	if err != nil {
		log.Print("err: Failed to load games")
	}

	// Start listening for gametimes from presences
	go counter.Listen()

	// Start gametime count for everyone already playing
	for _, server := range ready.Servers {
		for _, presence := range server.Presences {
			gameStarted(presence)
		}
	}

	log.Print("Everything set up")
}

func messageReceived(message discord.Message) {
	if !strings.HasPrefix(message.Content, "!go") {
		return
	}

	args := strings.Split(message.Content, " ")
	if len(args)-1 < 1 {
		return
	}

	totalCommands++

	command, ok := commands[args[1]]
	if ok {
		command.Handler(message, args...)
	} else {
		log.Printf("No command '%s'", args[1])
	}
}

func gameStarted(presence discord.Presence) {
	user := presence.GetUser(&client)
	game, gameExists := games[string(presence.GameID)]
	pUser, isPlaying := counter.InProgress[user.ID]

	if isPlaying && !gameExists {
		counter.GametimeChan <- pUser
	} else if isPlaying && gameExists {
		// User may be in more than one server with this instance of gobot
		return
	} else if !isPlaying && gameExists {
		counter.StartGametime(user, game)
	}
}

func getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}

func getUserCountString() string {
	users := 0
	channels := 0
	for _, server := range client.Servers {
		users += len(server.Members)
		channels += len(server.Channels)
	}
	return fmt.Sprintf(
		"%d in %d channels and %d servers",
		users,
		channels,
		len(client.Servers),
	)
}

func statsCommand(message discord.Message, args ...string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	client.SendMessage(
		message.ChannelID,
		fmt.Sprintf("Bot statistics:\n"+
			"`Uptime` %s\n"+
			"`Memory used` %.2f Mb\n"+
			"`Concurrent tasks` %d\n"+
			"`Users in touch` %s\n"+
			"`Users playing` %d\n"+
			"`Commands answered` %d\n",
			getDurationString(time.Now().Sub(startTime)),
			float64(stats.Alloc)/1000000,
			runtime.NumGoroutine(),
			getUserCountString(),
			len(counter.InProgress),
			totalCommands,
		),
	)
}

func helpCommand(message discord.Message, args ...string) {
	toSend := "Available commands:\n"
	for _, command := range commands {
		toSend += fmt.Sprintf("`%s` %s\n", command.Word, command.Help)
	}
	client.SendMessage(message.ChannelID, toSend)
}

func listRemindersCommand(message discord.Message) {
	list, err := reminder.GetUserReminders(message.Author)
	if err != nil {
		client.SendMessage(message.ChannelID, errorMessage)
		return
	}

	if len(list) < 1 {
		client.SendMessage(message.ChannelID, "You got no reminder :(")
		return
	}

	rString := "Your reminders :\n"
	for _, rem := range list {
		rString += fmt.Sprintf(
			"`%s` in `%s` -> %s\n",
			rem.UUID,
			getDurationString(rem.DurationLeft()),
			rem.Message,
		)
	}
	client.SendMessage(message.ChannelID, rString)
}

func reminderCommand(message discord.Message, args ...string) {
	if len(args)-1 < 2 {
		return
	}

	if args[2] == "list" {
		listRemindersCommand(message)
		return
	}
	// } else if args[2] == "delete" {
	// 	if len(args)-1 < 3 {
	// 		return
	// 	}
	// 	if err := reminder.RemoveReminder(message.Author, args[3]); err != nil {
	// 		client.SendMessage(message.ChannelID, errorMessage)
	// 	} else {
	// 		client.SendMessage(message.ChannelID, "Reminder deleted!")
	// 	}
	// 	return
	// }

	duration, err := time.ParseDuration(args[2])
	if err != nil {
		client.SendMessage(
			message.ChannelID,
			fmt.Sprintf("Couldn't understand that :("),
		)
		return
	}

	var reminderMessage string
	if len(args)-1 < 3 {
		reminderMessage = "`Reminder` ping !"
	} else {
		reminderMessage = fmt.Sprintf("`Reminder` %s", strings.Join(args[3:], " "))
	}

	client.SendMessage(
		message.ChannelID,
		fmt.Sprintf("Aight! I will ping you in %s.", duration.String()),
	)

	log.Printf("Reminding %s in %s", message.Author.Name, duration.String())
	reminder.NewReminder(message.Author, duration, reminderMessage)
}

func sourceCommand(message discord.Message, args ...string) {
	client.SendMessage(message.ChannelID, "Here you go! <https://github.com/gdraynz/go-discord>")
}

func voiceCommand(message discord.Message, args ...string) {
	if message.Author.Name != "steelou" {
		client.SendMessage(message.ChannelID, "Nah.")
		return
	}

	server := message.GetServer(&client)
	voiceChannel := client.GetChannel(server, "General")
	if err := client.SendAudio(voiceChannel, "Blue.mp3"); err != nil {
		log.Print(err)
	}
}

func playedCommand(message discord.Message, args ...string) {
	var pString string

	if len(args)-1 == 1 {
		userMap, err := counter.GetUserGametime(message.Author)
		if err != nil {
			pString = "I don't remember you playing anything I know :("
		} else {
			pString = "As far as I'm aware, you played:\n"
			for id, gametime := range userMap {
				pString += fmt.Sprintf(
					"`%s` %s\n",
					games[id].Name,
					getDurationString(time.Duration(gametime)),
				)
			}
		}
	} else if len(args)-1 == 2 && args[2] == "reset" {
		if err := counter.ResetGametime(message.Author); err != nil {
			pString = errorMessage
		} else {
			pString = "Pew! Everything gone!"
		}
	// } else if len(args)-1 == 2 && args[2] == "server" {
	// 	gameMap, err := counter.ServerGametime(message.GetServer(&client))
	// 	if err != nil {
	// 		log.Print(err)
	// 		pString = "I don't remember this server playing anything I know :("
	// 	} else {
	// 		pString = "This server played:\n"
	// 		for id, gametime := range gameMap {
	// 			pString += fmt.Sprintf(
	// 				"`%s` %s\n",
	// 				games[id].Name,
	// 				getDurationString(time.Duration(gametime)),
	// 			)
	// 		}
	// 	}
	} else if len(args)-1 >= 3 {
		gameName := strings.Join(args[3:], " ")
		log.Printf("resetting '%s'", gameName)
		game, err := getGameByName(gameName)
		if err != nil {
			pString = "I don't know that game :("
		} else {
			if err := counter.ResetOneGametime(message.Author, game); err != nil {
				pString = "Did you ever played that game ?"
			} else {
				pString = fmt.Sprintf("Pew! Game time for %s resetted!", game.Name)
			}
		}
	} else {
		pString = errorMessage
	}

	if _, err := client.SendMessage(message.ChannelID, pString); err != nil {
		log.Print(err)
	}
}

func twitchCommand(message discord.Message, args ...string) {
	var tString string

	if len(args)-1 < 2 {
		top, err := getTwitchTopGames()
		if err != nil {
			client.SendMessage(message.ChannelID, errorMessage)
			return
		}
		tString = "Here are the top 3 games streamed on Twitch:\n" + top
	} else {
		stream, err := getTwitchStream(strings.Join(args[2:], " "))
		if err != nil {
			client.SendMessage(message.ChannelID, errorMessage)
			return
		}
		tString = stream
	}

	client.SendMessage(message.ChannelID, tString)
}

func infoCommand(message discord.Message, args ...string) {
	var iString string

	iString = fmt.Sprintf(
		"Your informations:\n"+
			"`ID` %s\n"+
			"`Name` %s\n"+
			"`Avatar` <%s>",
		message.Author.ID,
		message.Author.Name,
		message.Author.AvatarURL(),
	)

	client.SendMessage(message.ChannelID, iString)
}

func main() {
	flag.Parse()

	// Logging
	var logfile *os.File
	if !*flagStdout {
		var err error
		logfile, err = os.OpenFile("gobot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logfile)
	}

	// Twitch client
	initTwitch()

	// Time counter
	counter, _ = NewCounter()

	// Time reminder
	reminder, _ = NewTimeReminder()

	client = discord.Client{
		OnReady:          onReady,
		OnMessageCreate:  messageReceived,
		OnPresenceUpdate: gameStarted,

		// Debug: true,
	}

	commands = map[string]Command{
		"help": Command{
			Word:    "help",
			Help:    "Prints the help message",
			Handler: helpCommand,
		},
		"reminder": Command{
			Word:    "reminder <XhYmZs> [<message>]",
			Help:    "Reminds you of something in X hours Y minutes Z seconds",
			Handler: reminderCommand,
		},
		"stats": Command{
			Word:    "stats",
			Help:    "Prints bot statistics",
			Handler: statsCommand,
		},
		"source": Command{
			Word:    "source",
			Help:    "Shows the bot's source URL",
			Handler: sourceCommand,
		},
		"played": Command{
			Word:    "played [reset [<game>]]",
			Help:    "Shows your game time",
			Handler: playedCommand,
		},
		"twitch": Command{
			Word:    "twitch [<game>]",
			Help:    "Show the top 3 streamed games on twitch, or link to the most watched stream for the given game",
			Handler: twitchCommand,
		},
		"info": Command{
			Word:    "info",
			Help:    "Shows the user informations",
			Handler: infoCommand,
		},
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down", sig)
		counter.Close()
		if logfile != nil {
			logfile.Close()
		}
		client.Stop()
		os.Exit(0)
	}(sigc)

	if err := client.LoginFromFile(*flagConf); err != nil {
		log.Fatal(err)
	}

	// log.Print(client.GetPrivateChannel(client.User))

	client.Run()
}
