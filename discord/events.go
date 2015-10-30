package discord

import (
	"time"
)

type ReadyEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   struct {
		HeartbeatInterval time.Duration `json:"heartbeat_interval"`
		User              User          `json:"user"`
		Servers           []Server      `json:"guilds"`
		PrivateChannels   []struct {
			ID string `json:"id"`
		} `json:"private_channels"`
	} `json:"d"`
}

type TypingEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   struct {
		UserID    string `json:"user_id"`
		Timestamp int    `json:"timestamp"`
		ChannelID string `json:"channel_id"`
	} `json:"d"`
}
