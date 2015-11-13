package discord

import (
	"encoding/json"
	"fmt"
)

// User defines a user of Disord
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
func (u *User) AvatarURL() string {
	if u.Avatar != "" {
		return fmt.Sprintf("%s/%s/avatars/%s.jpg", apiUsers, u.ID, u.Avatar)
	}
	return ""
}

// Ban bans the user from the given server
func (u *User) Ban(client *Client, server Server) error {
	return client.Ban(server, *u)
}

// Unban unbans the user from the given server
func (u *User) Unban(client *Client, server Server) error {
	return client.Unban(server, *u)
}

// Kick kicks the user from the given server
func (u *User) Kick(client *Client, server Server) error {
	return client.Kick(server, *u)
}

// CreatePrivateChannel creates a private channel with this user
func (u *User) CreatePrivateChannel(client *Client) (PrivateChannel, error) {
	return client.CreatePrivateChannel(*u)
}

// Presence defines the status of a User
type Presence struct {
	Status   string      `json:"status"`
	GameID   json.Number `json:"game_id,Number"`
	User     User        `json:"user"`
	ServerID string      `json:"guild_id"`
	Roles    []string    `json:"roles"`
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
