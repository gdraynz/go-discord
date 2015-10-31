package discord

import (
	"time"
)

type Ready struct {
	HeartbeatInterval time.Duration    `json:"heartbeat_interval"`
	User              User             `json:"user"`
	Servers           []Server         `json:"guilds"`
	PrivateChannels   []PrivateChannel `json:"private_channels"`
}

type readyEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   Ready  `json:"d"`
}
