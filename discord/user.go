package discord

import (
	"fmt"
)

type User struct {
	ID            string `json:"id"`
	Name          string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

func (u *User) GetAvatarURL() string {
	if u.Avatar != "" {
		return fmt.Sprintf("%s/avatars/%s.jpg", apiUsers, u.Avatar)
	}
	return ""
}
