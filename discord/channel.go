package discord

type Channel struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	Topic         string `json:"topic"`
	LastMessageID string `json:"last_message_id"`
	Type          string `json:"type"`
	Position      int    `json:"position"`
	ServerID      string `json:"guild_id"`
}

type channelCreateEvent struct {
	OpCode int     `json:"op"`
	Type   string  `json:"t"`
	Data   Channel `json:"d"`
}

type PrivateChannel struct {
	ID            string `json:"id"`
	Recipient     User   `json:"recipient"`
	LastMessageID string `json:"last_message_id"`
}

type privateChannelCreateEvent struct {
	OpCode int            `json:"op"`
	Type   string         `json:"t"`
	Data   PrivateChannel `json:"d"`
}
