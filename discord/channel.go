package discord

type Channel struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	Topic         string `json:"topic"`
	LastMessageID string `json:"last_message_id"`
	Type          string `json:"type"`
	Position      int    `json:"position"`
}

type PrivateChannel struct {
	ID            string `json:"id"`
	Recipient     User   `json:"recipient"`
	LastMessageID string `json:"last_message_id"`
}
