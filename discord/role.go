package discord

type Role struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Managed     bool   `json:"managed"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
}
