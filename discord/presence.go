package discord

type Presence struct {
	Status string `json:"status"`
	GameID int    `json:"game_id"`
	User   struct {
		ID string `json:"id"`
	} `json:"user"`
}
