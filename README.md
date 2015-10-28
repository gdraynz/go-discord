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
