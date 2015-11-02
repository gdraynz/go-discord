package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdraynz/go-discord/discord"
)

var (
	flagConf = flag.String("conf", "conf.json", "Configuration file")

	client discord.Client
)

func messageReceived(message discord.Message) {
	_, ok := client.PrivateChannels[message.ChannelID]
	if ok {
		switch message.Content {
		case "info":
			client.SendMessage(message.ChannelID, "Your name: "+message.Author.Name)
			client.SendMessage(message.ChannelID, "Your email: "+message.Author.Email)
		}
	}
}

func typingMessage(typing discord.Typing) {
	_, ok := client.PrivateChannels[typing.ChannelID]
	if ok {
		client.SendMessage(typing.ChannelID, "DON'T TALK TO ME")
	}
}

func main() {
	flag.Parse()

	client = discord.Client{
		OnMessageCreate: messageReceived,
		OnTypingStart:   typingMessage,
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
