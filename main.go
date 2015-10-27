package main

import (
	"log"

	"github.com/gdraynz/go-discord/discord"
)

func main() {
	c := discord.Client{}

	if err := c.LoginFromFile("conf.json"); err != nil {
		log.Fatal(err)
	}
}
