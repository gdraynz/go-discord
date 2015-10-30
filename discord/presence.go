package discord

type Presence struct {
	Status   string `json:"status"`
	GameID   int    `json:"game_id"`
	User     User   `json:"user"`
	ServerID string `json:"guild_id"`
	Roles    []Role `json:"roles"`
}
