package discord

// Channel defines everything about a channel
type Channel struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	Topic         string `json:"topic"`
	LastMessageID string `json:"last_message_id"`
	Type          string `json:"type"`
	Position      int    `json:"position"`
	ServerID      string `json:"guild_id"`
}

// SendMessage sends a message to the channel
func (channel *Channel) SendMessage(client *Client, content string) error {
	return client.SendMessage(channel.ID, content)
}

// SendMessage sends a message to the channel includind user mentions
func (channel *Channel) SendMessageMention(client *Client, content string, mentions []User) error {
	return client.SendMessageMention(channel.ID, content, mentions)
}

type channelCreateEvent struct {
	OpCode int     `json:"op"`
	Type   string  `json:"t"`
	Data   Channel `json:"d"`
}

// PrivateChannel defines everything about a private one-to-one conversation
type PrivateChannel struct {
	ID            string `json:"id"`
	Recipient     User   `json:"recipient"`
	LastMessageID string `json:"last_message_id"`
}

// SendMessage sends a message to the user linked to the PrivateChannel
func (private *PrivateChannel) SendMessage(client *Client, content string) error {
	return client.SendMessage(private.ID, content)
}

type privateChannelCreateEvent struct {
	OpCode int            `json:"op"`
	Type   string         `json:"t"`
	Data   PrivateChannel `json:"d"`
}
