package discord

type MessageEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   struct {
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
	} `json:"d"`
}
