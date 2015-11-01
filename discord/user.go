package discord

import (
	"fmt"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"username"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	Avatar   string `json:"avatar"`

	// TODO: Sometimes int, sometimes string.
	// Discriminator string `json:"discriminator,string"`
}

// GetAvatarURL returns the user's avatar URL
func (u *User) GetAvatarURL() string {
	if u.Avatar != "" {
		return fmt.Sprintf("%s/avatars/%s.jpg", apiUsers, u.Avatar)
	}
	return ""
}

// Ban bans the user from the given server
func (u *User) Ban(client Client, server Server) error {
	return client.Ban(server, *u)
}

// Unban unbans the user from the given server
func (u *User) Unban(client Client, server Server) error {
	return client.Unban(server, *u)
}

type Presence struct {
	Status   string `json:"status"`
	GameID   int    `json:"game_id"`
	User     User   `json:"user"`
	ServerID string `json:"guild_id"`
	Roles    []Role `json:"roles"`
}

type presenceEvent struct {
	OpCode int      `json:"op"`
	Type   string   `json:"t"`
	Data   Presence `json:"d"`
}
