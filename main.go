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
	cid := message.ChannelID

	_, ok := client.PrivateChannels[cid]
	if !ok {
		return
	}

	switch message.Content {
	case "info":
		client.SendMessage(cid, "Your name: "+message.Author.Name)
		client.SendMessage(cid, "Your email: "+message.Author.Email)
	}
}

func typingMessage(typing discord.Typing) {
	cid := typing.ChannelID

	_, ok := client.PrivateChannels[cid]

	if ok {
		client.SendMessage(cid, "DON'T TALK TO ME")
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
