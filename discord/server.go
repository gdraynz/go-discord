package discord

// Server defines everything server-related, including a list of members,
// a list of channels, the presence of each member...
type Server struct {
	Name         string     `json:"name"`
	ID           string     `json:"id"`
	OwnerID      string     `json:"owner_id"`
	Region       string     `json:"region"`
	AfkTimeout   int        `json:"afk_timeout"`
	AfkChannelID string     `json:"afk_channel_id"`
	Type         string     `json:"type"`
	Icon         string     `json:"icon"`
	JoinedAt     string     `json:"joined_at"`
	Large        bool       `json:"large"`
	Presences    []Presence `json:"presences"`
	Roles        []Role     `json:"roles"`
	Members      []Member   `json:"members`
	Channels     []Channel  `json:"channels"`
}

type serverEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   Server `json:"d"`
}

// Member defines a server member from the Ready event
type Member struct {
	User     User     `json:"user"`
	Roles    []string `json:"roles"`
	Muted    bool     `json:"mute"`
	Deafed   bool     `json:"deaf"`
	JoinedAt string   `json:"joined_at"`
	ServerID string   `json:"guild_id"`
}

type memberEvent struct {
	OpCode int    `json:"op"`
	Type   string `json:"t"`
	Data   Member `json:"d"`
}
