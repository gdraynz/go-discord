package discord

import "encoding/json"

// Presence defines the status of a User
type Presence struct {
	Status   string   `json:"status"`
	Game     Game     `json:"game"`
	User     User     `json:"user"`
	ServerID string   `json:"guild_id"`
	Roles    []string `json:"roles"`
}

// Game defines a game played in a presence update
type Game struct {
	Name string `json:"name"`
}

// GetUser returns the User object of this presence event
func (presence *Presence) GetUser(client *Client) User {
	return client.GetUserByID(presence.User.ID)
}

type presenceEvent struct {
	OpCode int      `json:"op"`
	Type   string   `json:"t"`
	Data   Presence `json:"d"`
}

type presenceUpdate struct {
	Game      Game       `json:"game"`
	IdleSince json.Token `json:"idle_since"`
}
