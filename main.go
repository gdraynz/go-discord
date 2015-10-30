package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdraynz/go-discord/discord"
)

func messageReceived(message discord.MessageEvent) {
	log.Printf("%s : %s",
		message.Data.Author.Name,
		message.Data.Content,
	)
}

func main() {
	client := discord.Client{
		OnMessageReceived: messageReceived,
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
