package main

import (
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

	counter *TimeCounter
)

type Command struct {
	Word    string
	Help    string
	Handler func(discord.Message, ...string)
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

func reminderCommand(message discord.Message, args ...string) {
	if len(args)-1 < 2 {
		return
	}

	duration, err := time.ParseDuration(args[2])
	if err != nil {
		client.SendMessage(
			message.ChannelID,
			fmt.Sprintf("Couldn't understand that :("),
		)
	} else {
		var reminderMessage string
		if len(args)-1 < 3 {
			reminderMessage = fmt.Sprintf("@%s ping !", message.Author.Name)
		} else {
			reminderMessage = fmt.Sprintf(
				"@%s %s !",
				message.Author.Name,
				strings.Join(args[3:], " "),
			)
		}
		client.SendMessage(
			message.ChannelID,
			fmt.Sprintf("Aight! I will ping you in %s.", duration.String()),
		)
		log.Printf("Reminding %s in %s", message.Author.Name, duration.String())
		time.AfterFunc(duration, func() {
			client.SendMessageMention(
				message.ChannelID,
				reminderMessage,
				[]discord.User{message.Author},
			)
		})
	}
}

func sourceCommand(message discord.Message, args ...string) {
	client.SendMessage(message.ChannelID, "https://github.com/gdraynz/go-discord")
}

func avatarCommand(message discord.Message, args ...string) {
	client.SendMessage(message.ChannelID, message.Author.GetAvatarURL())
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

	client.SendMessage(message.ChannelID, pString)
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

func main() {
	flag.Parse()

	// time counter
	counter, _ = NewCounter()

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
			Word:    "reminder <time [XhYmZs]> [<message>]",
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
		"avatar": Command{
			Word:    "avatar",
			Help:    "Shows your avatar URL",
			Handler: avatarCommand,
		},
		"played": Command{
			Word:    "played",
			Help:    "Shows your game time",
			Handler: playedCommand,
		},
		"twitch": Command{
			Word:    "twitch [<game>]",
			Help:    "Show the top 3 streamed games on twitch, or link to the most watched stream for the given game",
			Handler: twitchCommand,
		},
		// "watch": Command{
		// 	Word:    "watch <user> [<game>]",
		// 	Help:    "Ping you when <user> starts to play <game>",
		// 	Handler: watchCommand,
		// },
		// "unwatch": Command{
		// 	Word:    "unwatch <user> [<game>]",
		// 	Help:    "Stop notifying from the watch command",
		// 	Handler: unwatchCommand,
		// },
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

	client.Run()
}
