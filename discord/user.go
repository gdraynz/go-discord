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

func (u *User) GetAvatarURL() string {
	if u.Avatar != "" {
		return fmt.Sprintf("%s/avatars/%s.jpg", apiUsers, u.Avatar)
	}
	return ""
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
