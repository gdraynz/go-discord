package discord

import (
	"time"
)

type ReadyEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   struct {
		HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	} `json:"d"`
}

type Keepalive struct {
	OpCode int `json:"op"`
}
