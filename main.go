package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	Handler func(ChannelID string)
}

func onReady(ready discord.Ready) {
	startTime = time.Now()
}

func messageReceived(message discord.Message) {
	if !strings.HasPrefix(message.Content, "!go") {
		return
	}

	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		return
	}

	command, ok := commands[args[1]]
	if ok {
		command.Handler(message.ChannelID)
	} else {
		log.Printf("No command '%s'", args[1])
	}
}

func helpCommand(channelID string) {
	toSend := "Available commands:\n"
	for _, command := range commands {
		toSend += fmt.Sprintf("`%s` %s\n", command.Word, command.Help)
	}
	client.SendMessage(channelID, toSend)
}

func uptimeCommand(channelID string) {
	uptime := time.Now().Sub(startTime)
	toSend := fmt.Sprintf(
		"`Uptime` %0.2d:%02d:%02d",
		int(uptime.Hours()),
		int(uptime.Minutes()),
		int(uptime.Seconds()),
	)
	client.SendMessage(channelID, toSend)
}

func main() {
	flag.Parse()

	client = discord.Client{
		OnReady:         onReady,
		OnMessageCreate: messageReceived,
	}

	commands = map[string]Command{
		"uptime": Command{
			Word:    "uptime",
			Help:    "Shows the bot's uptime",
			Handler: uptimeCommand,
		},
		"help": Command{
			Word:    "help",
			Help:    "Prints the help message",
			Handler: helpCommand,
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
