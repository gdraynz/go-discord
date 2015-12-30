# Discord client for Go
<sup><sup>or The only Discord library that doesn't start with a D</sup></sup>

:warning: Please consider looking at [this library](https://github.com/bwmarrin/Discordgo), which is way more up-to-date.

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

### Go-bot

An [experimental bot](https://github.com/gdraynz/gobot) using this library.

### Todo

* Most of the audio functionalities have been worked on by many others libraries and integrated into Go by [discordgo](https://github.com/bwmarrin/dgvoice). I will try to integrate these into go-discord as soon as possible.

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
