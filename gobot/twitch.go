package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mrshankly/go-twitch/twitch"
)

var (
	twitchClient *twitch.Client
)

func initTwitch() {
	twitchClient = twitch.NewClient(&http.Client{})
}

func getGameURL(gameName string) string {
	return strings.Replace(
		fmt.Sprintf(
			"http://www.twitch.tv/directory/game/%s",
			gameName,
		), " ", "%20", -1)
}

func getTwitchStream(gameName string) (string, error) {
	var stream string

	live := true

	opt := &twitch.ListOptions{
		Limit:  1,
		Offset: 0,
		Game:   gameName,
		Live:   &live,
	}

	games, err := twitchClient.Streams.List(opt)
	if err != nil {
		return stream, err
	}

	if len(games.Streams)-1 < 0 {
		return stream, errors.New("No stream found")
	}

	game := games.Streams[0]

	stream = fmt.Sprintf(
		"**%s** on %s with %d viewers : <%s>\n",
		game.Channel.Name,
		game.Game,
		game.Viewers,
		game.Channel.Url,
	)

	return stream, nil
}

func getTwitchTopGames() (string, error) {
	var topGames string

	opt := &twitch.ListOptions{
		Limit:  3,
		Offset: 0,
	}

	games, err := twitchClient.Games.Top(opt)
	if err != nil {
		return topGames, err
	}

	for i, stream := range games.Top {
		topGames += fmt.Sprintf(
			"%d: **%s** with %d viewers\n",
			i+1,
			stream.Game.Name,
			stream.Viewers,
			// getGameURL(stream.Game.Name),
		)
	}

	return topGames, nil
}
