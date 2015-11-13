# Discord client for Go
<sup><sup>or The only Discord library that doesn't start with a D</sup></sup>

[![godoc badge](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gdraynz/go-discord/discord)
[![Go report](http://goreportcard.com/badge/gdraynz/go-discord)](http://goreportcard.com/report/gdraynz/go-discord)

API calls largely inspired by the Discord python client [discord.py](https://github.com/Rapptz/discord.py). (thanks)

Some methods may be pretty chaotic but I'm no Go expert and hope it will do for now.

## Basic usage

```go
package main

import (
    "log"

    "github.com/gdraynz/go-discord/discord"
)

func messageReceived(message discord.Message) {
    log.Printf("%s : %s",
        message.Author.Name,
        message.Content,
    )
}

func main() {
    client := discord.Client{
        OnMessageReceived: messageReceived,
    }

    if err := client.Login("email", "password"); err != nil {
        log.Fatal(err)
    }

    client.Run()
}
```

## Go-bot

### Current state

go-bot is the experiment that drives the developement of `go-discord`.
It currently handles some fun features :
* `!go played` : Gobot listen to each presence update and increment the playtime of users on its servers
* `!go reminder <time XhYmZs> [<message>]` : Set up a timer and simply ping the user after the specified time
* `!go twitch <game>` : Send the most watched twitch stream on the given game

### Todo

* I need useful ideas :(

## Related libraries

I'm not putting a lot of time on `go-discord`, here are many other implementations in different languages :

- [discord.py](https://github.com/Rapptz/discord.py) (Python)
- [discord.js](https://github.com/discord-js/discord.js) (JS)
- [discord.io](https://github.com/izy521/discord.io) (JS)
- [Discord.NET](https://github.com/RogueException/Discord.Net) (C#)
- [DiscordSharp](https://github.com/Luigifan/DiscordSharp) (C#)
- [Discord4J](https://github.com/knobody/Discord4J) (Java)
- [discordrb](https://github.com/meew0/discordrb) (Ruby)
- [Discordgo](https://github.com/bwmarrin/Discordgo) (Go)
- [discord](https://github.com/Xackery/discord) (Go)
