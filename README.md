# Discord Client for Go

[![godoc badge](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gdraynz/go-discord/discord)


Largely inspired by the Discord python client [discord.py](https://github.com/Rapptz/discord.py).

```golang
import "github.com/gdraynz/go-discord/discord"

c := discord.Client{}
if err := c.Login("email", "password"); err != nil {
    log.Fatal(err)
}

log.Print(c.token)
log.Print(c.gateway)
```

I'm not putting a lot of time on it, you can find other implementations here:

- [discord.py](https://github.com/Rapptz/discord.py)
- [discord.js](https://github.com/discord-js/discord.js)
- [discord.io](https://github.com/izy521/discord.io)
- [Discord.NET](https://github.com/RogueException/Discord.Net)
- [DiscordSharp](https://github.com/Luigifan/DiscordSharp)
- [Discord4J](https://github.com/knobody/Discord4J)
- [discordrb](https://github.com/meew0/discordrb)
