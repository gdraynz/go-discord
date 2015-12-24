package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	VERSION     = "v1.0.0"
	apiBase     = "https://discordapp.com/api"
	apiLogin    = apiBase + "/auth/login"
	apiLogout   = apiBase + "/auth/logout"
	apiRegister = apiBase + "/auth/register"
	apiChannels = apiBase + "/channels"
	apiGateway  = apiBase + "/gateway"
	apiServers  = apiBase + "/guilds"
	apiInvite   = apiBase + "/invite"
	apiUsers    = apiBase + "/users"
	apiVoice    = apiBase + "/voice"
)

// Client is the main object, instantiate it to use Discord Websocket API
type Client struct {
	OnReady                func(Ready)
	OnMessageCreate        func(Message)
	OnMessageAck           func(Message) // Only contains `id` and `channel_id`
	OnMessageUpdate        func(Message)
	OnMessageDelete        func(Message) // Only contains `id` and `channel_id`
	OnTypingStart          func(Typing)
	OnPresenceUpdate       func(Presence)
	OnChannelCreate        func(Channel)
	OnChannelUpdate        func(Channel)
	OnChannelDelete        func(Channel)
	OnPrivateChannelCreate func(PrivateChannel)
	OnPrivateChannelDelete func(PrivateChannel)
	OnServerCreate         func(Server)
	OnServerDelete         func(Server)
	OnServerMemberAdd      func(Member)
	OnServerMemberDelete   func(Member)

	// Reconnect upon websocket close server-side (EOF)
	Reconnect bool

	// Print websocket dumps (may be huge)
	Debug bool
	// Accessible, but you shouldn't modify these (I may put some getters there)
	User            User
	Servers         map[string]Server
	PrivateChannels map[string]PrivateChannel

	wsConn          *websocket.Conn
	gateway         gatewayStruct
	token           tokenStruct
	keepaliveTicker *time.Ticker
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{}

	req.Header.Set(
		"User-Agent",
		fmt.Sprintf("DiscordBot (https://github.com/gdraynz/go-discord, %s)", VERSION),
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// JSON from payload
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("%s %s : %s", req.Method, req.URL.String, string(body[:]))
	}

	return body, err
}

func (c *Client) doHandshake() {
	log.Print("Sending handshake")
	c.wsConn.WriteJSON(map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token": c.token.Value,
			"properties": map[string]string{
				"$os":               "linux",
				"$browser":          "go-discord",
				"$device":           "go-discord",
				"$referer":          "",
				"$referring_domain": "",
			},
			"v": 3,
		},
	})
}

func (c *Client) initServers(ready Ready) {
	c.Servers = make(map[string]Server)
	c.PrivateChannels = make(map[string]PrivateChannel)
	for _, server := range ready.Servers {
		// Set ServerID of each channel
		for i := range server.Channels {
			// https://github.com/golang/go/issues/3117
			tmp := server.Channels[i]
			tmp.ServerID = server.ID
			server.Channels[i] = tmp
		}
		c.Servers[server.ID] = server
	}
	for _, private := range ready.PrivateChannels {
		c.PrivateChannels[private.ID] = private
	}
}

func (c *Client) handleReady(eventStr []byte) {
	var ready readyEvent
	if err := json.Unmarshal(eventStr, &ready); err != nil {
		log.Printf("handleReady: %s", err)
		return
	}

	// WebSocket keepalive
	go func() {
		c.keepaliveTicker = time.NewTicker(ready.Data.HeartbeatInterval * time.Millisecond)
		for range c.keepaliveTicker.C {
			timestamp := int(time.Now().Unix())
			log.Printf("Sending keepalive with timestamp %d", timestamp)
			c.wsConn.WriteJSON(map[string]int{
				"op": 1,
				"d":  timestamp,
			})
		}
	}()

	c.User = ready.Data.User
	c.initServers(ready.Data)

	if c.OnReady == nil {
		if c.Debug {
			log.Print("No handler for READY")
		}
	} else {
		log.Print("Client ready, calling OnReady handler")
		c.OnReady(ready.Data)
	}
}

func (c *Client) handleMessageCreate(eventStr []byte) {
	if c.OnMessageCreate == nil {
		if c.Debug {
			log.Print("No handler for MESSAGE_CREATE")
		}
		return
	}

	var message messageEvent
	if err := json.Unmarshal(eventStr, &message); err != nil {
		log.Printf("messageCreate: %s", err)
		return
	}

	if message.Data.Author.ID != c.User.ID {
		c.OnMessageCreate(message.Data)
	} else {
		log.Print("Ignoring message from self")
	}
}

func (c *Client) handleMessageAck(eventStr []byte) {
	if c.OnMessageAck == nil {
		if c.Debug {
			log.Print("No handler for MESSAGE_ACK")
		}
		return
	}

	var message messageEvent
	if err := json.Unmarshal(eventStr, &message); err != nil {
		log.Printf("messageAck: %s", err)
		return
	}

	c.OnMessageAck(message.Data)
}

func (c *Client) handleMessageUpdate(eventStr []byte) {
	if c.OnMessageUpdate == nil {
		if c.Debug {
			log.Print("No handler for MESSAGE_UPDATE")
		}
		return
	}

	var message messageEvent
	if err := json.Unmarshal(eventStr, &message); err != nil {
		log.Printf("messageUpdate: %s", err)
		return
	}

	if message.Data.Author.ID != c.User.ID {
		c.OnMessageUpdate(message.Data)
	} else {
		log.Print("Ignoring updated message from self")
	}
}

func (c *Client) handleMessageDelete(eventStr []byte) {
	if c.OnMessageDelete == nil {
		if c.Debug {
			log.Print("No handler for MESSAGE_DELETE")
		}
		return
	}

	var message messageEvent
	if err := json.Unmarshal(eventStr, &message); err != nil {
		log.Printf("messageDelete: %s", err)
		return
	}

	c.OnMessageDelete(message.Data)
}

func (c *Client) handleTypingStart(eventStr []byte) {
	if c.OnTypingStart == nil {
		if c.Debug {
			log.Print("No handler for TYPING_START")
		}
		return
	}

	var typing typingEvent
	if err := json.Unmarshal(eventStr, &typing); err != nil {
		log.Printf("typingStart: %s", err)
		return
	}

	c.OnTypingStart(typing.Data)
}

func (c *Client) handlePresenceUpdate(eventStr []byte) {
	if c.OnPresenceUpdate == nil {
		if c.Debug {
			log.Print("No handler for PRESENCE_UPDATE")
		}
		return
	}

	var presence presenceEvent
	if err := json.Unmarshal(eventStr, &presence); err != nil {
		log.Printf("presenceUpdate: %s", err)
		return
	}

	c.OnPresenceUpdate(presence.Data)
}

func (c *Client) handleChannelCreate(eventStr []byte) {
	var channelCreate interface{}
	if err := json.Unmarshal(eventStr, &channelCreate); err != nil {
		log.Printf("handleChannelCreate: %s", err)
		return
	}

	// woot
	isPrivate := channelCreate.(map[string]interface{})["d"].(map[string]interface{})["is_private"].(bool)

	if isPrivate {
		var event privateChannelEvent
		if err := json.Unmarshal(eventStr, &event); err != nil {
			log.Printf("privateChannelCreate: %s", err)
			return
		}

		privateChannel := event.Data
		c.PrivateChannels[privateChannel.ID] = privateChannel

		if c.OnPrivateChannelCreate == nil {
			if c.Debug {
				log.Print("No handler for private CHANNEL_CREATE")
			}
		} else {
			c.OnPrivateChannelCreate(privateChannel)
		}
	} else {
		var event channelEvent
		if err := json.Unmarshal(eventStr, &event); err != nil {
			log.Printf("channelCreate: %s", err)
			return
		}

		channel := event.Data
		// XXX: Workaround for c.Channels[private.ID].Private = true
		// https://github.com/golang/go/issues/3117
		tmp := c.Servers[channel.ServerID]
		tmp.Channels = append(tmp.Channels, channel)
		c.Servers[channel.ServerID] = tmp

		if c.OnChannelCreate == nil {
			if c.Debug {
				log.Print("No handler for CHANNEL_CREATE")
			}
		} else {
			c.OnChannelCreate(channel)
		}
	}
}

func (c *Client) handleChannelUpdate(eventStr []byte) {
	var event channelEvent
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Printf("channelUpdate: %s", err)
		return
	}

	channel := event.Data
	// Get channel id in slice of server
	i := c.getChannelIndex(channel.ID)
	// XXX: Workaround for c.Servers[channel.ServerID].Channels = ...
	// https://github.com/golang/go/issues/3117
	tmp := c.Servers[channel.ServerID]
	tmp.Channels = append(tmp.Channels[:i], tmp.Channels[i+1:]...)
	tmp.Channels = append(tmp.Channels, channel)
	c.Servers[channel.ServerID] = tmp

	if c.OnChannelUpdate == nil {
		if c.Debug {
			log.Print("No handler for CHANNEL_UPDATE")
		}
	} else {
		c.OnChannelUpdate(channel)
	}
}

func (c *Client) handleChannelDelete(eventStr []byte) {
	var channelDelete interface{}
	if err := json.Unmarshal(eventStr, &channelDelete); err != nil {
		log.Printf("handleChannelDelete: %s", err)
		return
	}

	// woot #2
	isPrivate := channelDelete.(map[string]interface{})["d"].(map[string]interface{})["is_private"].(bool)

	if isPrivate {
		var event privateChannelEvent
		if err := json.Unmarshal(eventStr, &event); err != nil {
			log.Printf("privateChannelCreate: %s", err)
			return
		}

		privateChannel := event.Data
		delete(c.PrivateChannels, privateChannel.ID)

		if c.OnPrivateChannelDelete == nil {
			if c.Debug {
				log.Print("No handler for private CHANNEL_DELETE")
			}
		} else {
			c.OnPrivateChannelDelete(privateChannel)
		}
	} else {
		var event channelEvent
		if err := json.Unmarshal(eventStr, &event); err != nil {
			log.Printf("channelDelete: %s", err)
			return
		}

		channel := event.Data
		// Get channel id in slice of server
		i := c.getChannelIndex(channel.ID)
		// XXX: Workaround for c.Servers[channel.ServerID].Channels = ...
		// https://github.com/golang/go/issues/3117
		tmp := c.Servers[channel.ServerID]
		tmp.Channels = append(tmp.Channels[:i], tmp.Channels[i+1:]...)
		c.Servers[channel.ServerID] = tmp

		if c.OnChannelDelete == nil {
			if c.Debug {
				log.Print("No handler for CHANNEL_DELETE")
			}
		} else {
			c.OnChannelDelete(channel)
		}
	}
}

func (c *Client) handleGuildCreate(eventStr []byte) {
	var event serverEvent
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Printf("guildCreate: %s", err)
		return
	}

	server := event.Data
	c.Servers[server.ID] = server

	if c.OnServerCreate == nil {
		if c.Debug {
			log.Print("No handler for GUILD_CREATE")
		}
	} else {
		c.OnServerCreate(server)
	}
}

func (c *Client) handleGuildDelete(eventStr []byte) {
	var event serverEvent
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Printf("guildDelete: %s", err)
		return
	}

	server := event.Data
	delete(c.Servers, server.ID)

	if c.OnServerDelete == nil {
		if c.Debug {
			log.Print("No handler for GUILD_DELETE")
		}
	} else {
		c.OnServerDelete(server)
	}
}

func (c *Client) handleGuildMemberAdd(eventStr []byte) {
	var event memberEvent
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Printf("guildMemberAdd: %s", err)
		return
	}

	member := event.Data
	// https://github.com/golang/go/issues/3117
	tmp := c.Servers[member.ServerID]
	tmp.Members = append(tmp.Members, member)
	c.Servers[member.ServerID] = tmp

	if c.OnServerMemberAdd == nil {
		if c.Debug {
			log.Print("No handler for GUILD_MEMBER_ADD")
		}
	} else {
		c.OnServerMemberAdd(member)
	}
}

func (c *Client) handleGuildMemberDelete(eventStr []byte) {
	var event memberEvent
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Printf("guildMemberDelete: %s", err)
		return
	}

	member := event.Data
	// Get member id in slice of server
	i := c.getMemberIndex(member.User.ID)
	// https://github.com/golang/go/issues/3117
	tmp := c.Servers[member.ServerID]
	tmp.Members = append(tmp.Members[:i], tmp.Members[i+1:]...)
	c.Servers[member.ServerID] = tmp

	if c.OnServerMemberDelete == nil {
		if c.Debug {
			log.Print("No handler for GUILD_MEMBER_DELETE")
		}
	} else {
		c.OnServerMemberDelete(member)
	}
}

func (c *Client) handleEvent(eventStr []byte) {
	var event interface{}
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Print(err)
		return
	}

	eventType := event.(map[string]interface{})["t"].(string)

	if c.Debug {
		log.Printf("%s : %s", eventType, string(eventStr[:]))
	}

	// TODO: There must be a better way to directly cast the eventStr
	// to its corresponding object, avoiding double-unmarshal
	switch eventType {
	case "READY":
		c.handleReady(eventStr)
	case "MESSAGE_CREATE":
		c.handleMessageCreate(eventStr)
	case "MESSAGE_ACK":
		c.handleMessageAck(eventStr)
	case "MESSAGE_UPDATE":
		c.handleMessageUpdate(eventStr)
	case "MESSAGE_DELETE":
		c.handleMessageDelete(eventStr)
	case "TYPING_START":
		c.handleTypingStart(eventStr)
	case "PRESENCE_UPDATE":
		c.handlePresenceUpdate(eventStr)
	case "CHANNEL_CREATE":
		c.handleChannelCreate(eventStr)
	case "CHANNEL_UPDATE":
		c.handleChannelUpdate(eventStr)
	case "CHANNEL_DELETE":
		c.handleChannelDelete(eventStr)
	case "GUILD_CREATE":
		c.handleGuildCreate(eventStr)
	case "GUILD_DELETE":
		c.handleGuildDelete(eventStr)
	case "GUILD_MEMBER_ADD":
		c.handleGuildMemberAdd(eventStr)
	case "GUILD_MEMBER_DELETE":
		c.handleGuildMemberDelete(eventStr)
	default:
		if c.Debug {
			log.Printf("Ignoring %s", eventType)
			log.Printf("event dump: %s", string(eventStr[:]))
		}
	}

}

func (c *Client) getChannelIndex(channelID string) int {
	var channelPos int
	for _, server := range c.Servers {
		for i, channel := range server.Channels {
			if channel.ID == channelID {
				channelPos = i
				break
			}
		}
	}
	return channelPos
}

func (c *Client) getMemberIndex(memberID string) int {
	var memberPos int
	for _, server := range c.Servers {
		for i, member := range server.Members {
			if member.User.ID == memberID {
				memberPos = i
				break
			}
		}
	}
	return memberPos
}

// Get sends a GET request to the given url
func (c *Client) get(url string) ([]byte, error) {
	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.token.Value)

	return c.doRequest(req)
}

// Post sends a POST request with payload to the given url
func (c *Client) request(method string, url string, payload interface{}) ([]byte, error) {
	payloadJSON, _ := json.Marshal(payload)
	contentReader := bytes.NewReader(payloadJSON)

	// Prepare request
	req, err := http.NewRequest(method, url, contentReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.token.Value)
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// Login initialize Discord connection by requesting a token
func (c *Client) Login(email string, password string) error {
	// Prepare POST json
	m := map[string]string{
		"email":    email,
		"password": password,
	}

	// Get token
	tokenResp, err := c.request("POST", apiLogin, m)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(tokenResp, &c.token); err != nil {
		return err
	}

	// Get websocket gateway
	gatewayResp, err := c.get(apiGateway)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(gatewayResp, &c.gateway); err != nil {
		return err
	}

	return nil
}

// LoginFromFile call login with email and password found in the given file
func (c *Client) LoginFromFile(filename string) error {
	fileDump, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	type fileCredentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var creds = fileCredentials{}
	if err := json.Unmarshal(fileDump, &creds); err != nil {
		return err
	}

	return c.Login(creds.Email, creds.Password)
}

// SendPresence set the given game name as playing to discord
func (c *Client) SendPresence(game string) error {
	log.Printf("Sending presence for game %s", game)
	data := map[string]interface{}{
		"op": 3,
		"d": presenceUpdate{
			Game: Game{
				Name: game,
			},
			IdleSince: nil,
		},
	}
	return c.wsConn.WriteJSON(data)
}

// GetChannel returns the Channel object from the given channel name on the given server name
func (c *Client) GetChannel(server Server, channelName string) Channel {
	var res Channel
	for _, channel := range server.Channels {
		if channel.Name == channelName {
			res = channel
			break
		}
	}
	return res
}

// GetChannelByID returns the Channel object from the given ID as well as its position in its server's channel list
func (c *Client) GetChannelByID(channelID string) Channel {
	var res Channel
	for _, server := range c.Servers {
		for _, channel := range server.Channels {
			if channel.ID == channelID {
				res = channel
				break
			}
		}
	}
	return res
}

// GetServer returns the Server object from the given server name
func (c *Client) GetServer(serverName string) Server {
	var res Server
	for _, server := range c.Servers {
		if server.Name == serverName {
			res = server
			break
		}
	}
	return res
}

// GetUser returns the User object on the specified server using the given name
func (c *Client) GetUser(server Server, userName string) User {
	var res User
	for _, member := range server.Members {
		if member.User.Name == userName {
			res = member.User
			break
		}
	}
	return res
}

// GetUserByID returns the User object from the given user ID
func (c *Client) GetUserByID(userID string) User {
	var res User
	for _, server := range c.Servers {
		for _, member := range server.Members {
			if member.User.ID == userID {
				res = member.User
				break
			}
		}
	}
	return res
}

// JoinServer receive an invite ID and tries to join the corresponding server/channel
func (c *Client) JoinServer(inviteID string) error {
	_, err := c.request(
		"POST",
		fmt.Sprintf("%s/%s", apiInvite, inviteID),
		nil,
	)
	return err
}

// SendMessage sends a message to the given channel
// XXX: string sent as channel ID because of Channel/PrivateChannel differences
func (c *Client) SendMessage(channelID string, content string) (Message, error) {
	var message Message

	response, err := c.request(
		"POST",
		fmt.Sprintf(apiChannels+"/%s/messages", channelID),
		map[string]string{
			"content": content,
		},
	)
	if err != nil {
		return message, err
	}

	if err := json.Unmarshal(response, &message); err != nil {
		return message, err
	}

	return message, err
}

// SendMessageMention sends a message to the given channel mentionning users
// XXX: string sent as channel ID because of Channel/PrivateChannel differences
func (c *Client) SendMessageMention(channelID string, content string, mentions []User) (Message, error) {
	var message Message

	var userMentions []string
	for _, user := range mentions {
		userMentions = append(userMentions, user.ID)
	}

	response, err := c.request(
		"POST",
		fmt.Sprintf(apiChannels+"/%s/messages", channelID),
		map[string]interface{}{
			"content":  content,
			"mentions": userMentions,
		},
	)

	if err := json.Unmarshal(response, &message); err != nil {
		return message, err
	}

	return message, err
}

// GetPrivateChannel returns the private channel corresponding to the user
func (c *Client) GetPrivateChannel(user User) (pc PrivateChannel) {
	found := false

	for _, private := range c.PrivateChannels {
		if private.Recipient.ID == user.ID {
			pc = private
			found = true
			break
		}
	}

	if !found {
		pc, _ = c.CreatePrivateChannel(user)
	}

	return pc
}

// CreatePrivateChannel creates a private channel with the given user
func (c *Client) CreatePrivateChannel(user User) (PrivateChannel, error) {
	var pChannel PrivateChannel

	response, err := c.request(
		"POST",
		fmt.Sprintf("%s/%s/channels", apiUsers, c.User.ID),
		map[string]string{
			"recipient_id": user.ID,
		},
	)

	if err != nil {
		return pChannel, err
	}

	if err := json.Unmarshal(response, &pChannel); err != nil {
		return pChannel, err
	}

	c.PrivateChannels[pChannel.ID] = pChannel

	return pChannel, err
}

// AckMessage acknowledges the message on the given channel
func (c *Client) AckMessage(channel Channel, message Message) error {
	_, err := c.request(
		"POST",
		fmt.Sprintf("%s/%s/messages/%s/ack", apiChannels, channel.ID, message.ID),
		nil,
	)
	return err
}

// EditMessage modifies the message from the channel with the given ID.
// It takes a new content string and a list of mentions.
func (c *Client) EditMessage(channelID string, messageID string, content string) (Message, error) {
	var message Message

	response, err := c.request(
		"PATCH",
		fmt.Sprintf("%s/%s/messages/%s", apiChannels, channelID, messageID),
		map[string]interface{}{
			"content": content,
		},
	)
	if err != nil {
		return message, err
	}

	if err := json.Unmarshal(response, &message); err != nil {
		return message, err
	}

	return message, err
}

// DeleteMessage deletes the message from the channel with the given ID
func (c *Client) DeleteMessage(channel Channel, message Message) error {
	_, err := c.request(
		"DELETE",
		fmt.Sprintf("%s/%s/messages/%s", apiChannels, channel.ID, message.ID),
		nil,
	)
	return err
}

// Ban bans a user from the giver server
func (c *Client) Ban(server Server, user User) error {
	_, err := c.request(
		"PUT",
		fmt.Sprintf("%s/%s/bans/%s", apiServers, server.ID, user.ID),
		nil,
	)
	return err
}

// Unban unbans a user from the giver server
func (c *Client) Unban(server Server, user User) error {
	_, err := c.request(
		"DELETE",
		fmt.Sprintf("%s/%s/bans/%s", apiServers, server.ID, user.ID),
		nil,
	)
	return err
}

// Kick kicks a user from the giver server
func (c *Client) Kick(server Server, user User) error {
	_, err := c.request(
		"DELETE",
		fmt.Sprintf("%s/%s/members/%s", apiServers, server.ID, user.ID),
		nil,
	)
	return err
}

// CreateChannel creates a new channel in the given server
func (c *Client) CreateChannel(server Server, name string, channelType string) error {
	_, err := c.request(
		"POST",
		fmt.Sprintf("%s/%s/channels", apiServers, server.ID),
		map[string]string{
			"name": name,
			"type": channelType,
		},
	)
	return err
}

// EditChannel edits a channel with the given parameters
// among (name string, topic string, position int)
func (c *Client) EditChannel(channel Channel, params map[string]interface{}) error {
	_, err := c.request(
		"PATCH",
		fmt.Sprintf("%s/%s", apiChannels, channel.ID),
		params,
	)
	return err
}

// GetRegion returns the Region object corresponding to the given server
func (c *Client) GetRegion(server Server) (Region, error) {
	var region Region

	response, err := c.get(apiVoice + "/regions")
	if err != nil {
		return region, err
	}

	var regions []Region
	if err := json.Unmarshal(response, &regions); err != nil {
		return region, err
	}

	for _, r := range regions {
		if r.ID == server.Region {
			region = r
			break
		}
	}

	return region, nil
}

// Run init the WebSocket connection and starts listening on it
func (c *Client) Run() {
	log.Printf("Setting up websocket to %s", c.gateway.Value)
	conn, _, err := websocket.DefaultDialer.Dial(c.gateway.Value, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Print("Connected")
	c.wsConn = conn

	for i := 0; i < 1 || c.Reconnect; i++ {
		c.doHandshake()

		for {
			_, message, err := c.wsConn.ReadMessage()
			if err != nil {
				log.Print(err)
				c.keepaliveTicker.Stop()
				break
			}
			go c.handleEvent(message)
		}
	}
}

// Stop closes the WebSocket connection
func (c *Client) Stop() {
	log.Print("Closing connection")
	c.keepaliveTicker.Stop()
	c.wsConn.Close()
}
