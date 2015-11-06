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

var (
	flagConf = flag.String("conf", "conf.json", "Configuration file")

	client    discord.Client
	startTime time.Time
	commands  map[string]Command
)

type Command struct {
	Word    string
	Help    string
	Handler func(discord.Message, ...string)
}

func onReady(ready discord.Ready) {
	startTime = time.Now()
}

func messageReceived(message discord.Message) {
	if !strings.HasPrefix(message.Content, "!go") {
		return
	}

	args := strings.Split(message.Content, " ")
	if len(args)-1 < 1 {
		return
	}

	command, ok := commands[args[1]]
	if ok {
		command.Handler(message, args...)
	} else {
		log.Printf("No command '%s'", args[1])
	}
}

func getUptime() string {
	uptime := time.Now().Sub(startTime)
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(uptime.Hours()),
		int(uptime.Minutes())%60,
		int(uptime.Seconds())%60,
	)
}

func statsCommand(message discord.Message, args ...string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	client.SendMessage(
		message.ChannelID,
		fmt.Sprintf("Bot statistics:\n"+
			"`Memory used` %.2f Mb\n"+
			"`Number of servers` %d\n"+
			"`Uptime` %s",
			float64(stats.Alloc)/1000000, len(client.Servers), getUptime(),
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
		log.Printf("Remind %s in %s", message.Author.Name, duration.String())
		time.AfterFunc(duration, func() {
			client.SendMessageMention(
				message.ChannelID,
				reminderMessage,
				[]discord.User{message.Author},
			)
		})
	}
}

func main() {
	flag.Parse()

	client = discord.Client{
		OnReady:         onReady,
		OnMessageCreate: messageReceived,
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
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down", sig)
		client.Stop()
		os.Exit(0)
	}(sigc)

	if err := client.LoginFromFile(*flagConf); err != nil {
		log.Fatal(err)
	}

	client.Run()
}
