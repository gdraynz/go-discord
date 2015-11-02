# Discord client for Go
<sup><sup>or The only Discord library that doesn't start with a D</sup></sup>

[![godoc badge](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gdraynz/go-discord/discord)

API calls largely inspired by the Discord python client [discord.py](https://github.com/Rapptz/discord.py). (thanks)

Some methods may be pretty chaotic but I'm no Go expert and hope it will do for now.

An usage example can be found in the `main.go` file.

```go
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

I'm not putting a lot of time on it, here are many other implementations in different languages :

- [discord.py](https://github.com/Rapptz/discord.py)
- [discord.js](https://github.com/discord-js/discord.js)
- [discord.io](https://github.com/izy521/discord.io)
- [Discord.NET](https://github.com/RogueException/Discord.Net)
- [DiscordSharp](https://github.com/Luigifan/DiscordSharp)
- [Discord4J](https://github.com/knobody/Discord4J)
- [discordrb](https://github.com/meew0/discordrb)
