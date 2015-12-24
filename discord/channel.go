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

	// Don't know how its formatted yet
	PermissionOverwrites []interface{} `json:"permission_overwrites"`
}

func (channel *Channel) GetServer(client *Client) Server {
	return client.Servers[channel.ServerID]
}

// SendMessage sends a message to the channel
func (channel *Channel) SendMessage(client *Client, content string) (Message, error) {
	return client.SendMessage(channel.ID, content)
}

type channelEvent struct {
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
func (private *PrivateChannel) SendMessage(client *Client, content string) (Message, error) {
	return client.SendMessage(private.ID, content)
}

type privateChannelEvent struct {
	OpCode int            `json:"op"`
	Type   string         `json:"t"`
	Data   PrivateChannel `json:"d"`
}
