package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdraynz/go-discord/discord"
)

var client discord.Client

func messageReceived(message discord.MessageEvent) {
	log.Printf("%s : %s",
		message.Data.Author.Name,
		message.Data.Content,
	)

	cid := message.Data.ChannelID
	if client.Channels[cid].Private {
		err := client.SendMessage(cid, "")
		if err != nil {
			log.Print(err)
		}
	}
}

func typingMessage(typing discord.TypingEvent) {
	cid := typing.Data.ChannelID
	if client.Channels[cid].Private {
		err := client.SendMessage(cid, "DONT TALK TO ME")
		if err != nil {
			log.Print(err)
		}
	}
}

func main() {
	client = discord.Client{
		OnMessageCreate: messageReceived,
		OnTypingStart:   typingMessage,
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		client.Stop()
		os.Exit(0)
	}(sigc)

	if err := client.LoginFromFile("conf.json"); err != nil {
		log.Fatal(err)
	}

	client.Run()
}
