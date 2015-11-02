package discord

// Role defines a role on a Discord server
type Role struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Managed     bool   `json:"managed"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
}
