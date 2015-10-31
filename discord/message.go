package discord

type Message struct {
	EditedTimestamp string `json:"edited_timestamp"`
	Timestanmp      string `json:"timestamp"`
	TTS             bool   `json:"tts"`
	Content         string `json:"content"`
	MentionEveryone bool   `json:"mention_everyone"`
	ID              string `json:"id"`
	ChannelID       string `json:"channel_id"`
	Author          User   `json:"author"`
	Mentions        []User `json:"mentions"`

	// TODO: Don't know how these are typed
	Attachments interface{} `json:"attachments"`
	Embeds      interface{} `json:"embeds"`
}

type messageEvent struct {
	OpCode int     `json:"op"`
	Type   string  `json:"t"`
	Data   Message `json:"d"`
}

type Typing struct {
	UserID    string `json:"user_id"`
	Timestamp int    `json:"timestamp"`
	ChannelID string `json:"channel_id"`
}

type typingEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   Typing `json:"d"`
}
