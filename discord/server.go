package discord

// Server defines everything server-related, including a list of members,
// a list of channels, the presence of each member...
type Server struct {
	Name         string     `json:"name"`
	ID           string     `json:"id"`
	OwnerID      string     `json:"owner_id"`
	Roles        []Role     `json:"roles"`
	Region       string     `json:"region"`
	AfkTimeout   int        `json:"afk_timeout"`
	AfkChannelID string     `json:"afk_channel_id"`
	Members      []Member   `json:"members`
	Type         string     `json:"type"`
	Channels     []Channel  `json:"channels"`
	Icon         string     `json:"icon"`
	JoinedAt     string     `json:"joined_at"`
	Large        bool       `json:"large"`
	Presences    []Presence `json:"presences"`
}

// Member defines a server member from the Ready event
type Member struct {
	User     User   `json:"user"`
	Roles    []Role `json:"roles"`
	Muted    bool   `json:"mute"`
	Deafed   bool   `json:"deaf"`
	JoinedAt string `json:"joined_at"`
}
