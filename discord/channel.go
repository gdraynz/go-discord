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

func (channel *Channel) SendMessage(client *Client, content string) error {
	return client.SendMessage(channel.ID, content)
}

func (channel *Channel) SendMessageMention(client *Client, content string, mentions []User) error {
	return client.SendMessageMention(channel.ID, content, mentions)
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

func (private *PrivateChannel) SendMessage(client *Client, content string) error {
	return client.SendMessage(private.ID, content)
}

type privateChannelCreateEvent struct {
	OpCode int            `json:"op"`
	Type   string         `json:"t"`
	Data   PrivateChannel `json:"d"`
}
