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
	apiBase     = "https://discordapp.com/api"
	apiGateway  = apiBase + "/gateway"
	apiUsers    = apiBase + "/users"
	apiRegister = apiBase + "/auth/register"
	apiLogin    = apiBase + "/auth/login"
	apiLogout   = apiBase + "/auth/logout"
	apiServers  = apiBase + "/guilds"
	apiChannels = apiBase + "/channels"
)

type Client struct {
	OnReady         func(ReadyEvent)
	OnMessageCreate func(MessageEvent)
	OnTypingStart   func(TypingEvent)

	Channels map[string]Channel

	user    User
	wsConn  *websocket.Conn
	gateway string
	token   string
}

func do_request(req *http.Request) (interface{}, error) {
	client := &http.Client{}
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

	log.Printf("response: %s", string(body[:]))
	var reqResult = map[string]interface{}{}
	if err := json.Unmarshal(body, &reqResult); err != nil {
		return nil, err
	}

	return reqResult, nil
}

func (c *Client) doHandshake() {
	log.Print("Sending handshake")
	c.wsConn.WriteJSON(map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token": c.token,
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

func (c *Client) initChannels(ready ReadyEvent) {
	c.Channels = make(map[string]Channel)
	for _, server := range ready.Data.Servers {
		for _, channel := range server.Channels {
			c.Channels[channel.ID] = channel
		}
		for _, private := range ready.Data.PrivateChannels {
			// XXX: Workaround for c.Channels[private.ID].Private = true
			// https://github.com/golang/go/issues/3117
			var tmp = c.Channels[private.ID]
			tmp.Private = true
			c.Channels[private.ID] = tmp
		}
	}
}

func (c *Client) handleReady(eventStr []byte) {
	var ready ReadyEvent
	if err := json.Unmarshal(eventStr, &ready); err != nil {
		log.Printf("handleReady: %s", err)
		return
	}

	go func() {
		ticker := time.NewTicker(ready.Data.HeartbeatInterval * time.Millisecond)
		for range ticker.C {
			timestamp := int(time.Now().Unix())
			log.Printf("Sending keepalive with timestamp %d", timestamp)
			c.wsConn.WriteJSON(map[string]int{
				"op": 1,
				"d":  timestamp,
			})
		}
	}()

	c.user = ready.Data.User
	c.initChannels(ready)

	if c.OnReady == nil {
		log.Print("No handler for READY")
	} else {
		c.OnReady(ready)
	}
}

func (c *Client) handleMessageCreate(eventStr []byte) {
	if c.OnMessageCreate == nil {
		log.Print("No handler for MESSAGE_CREATE")
		return
	}

	var message MessageEvent
	if err := json.Unmarshal(eventStr, &message); err != nil {
		log.Printf("messageCreate: %s", err)
		return
	}

	if message.Data.Author.ID != c.user.ID {
		c.OnMessageCreate(message)
	} else {
		log.Print("Ignoring message from self")
	}
}

func (c *Client) handleTypingStart(eventStr []byte) {
	if c.OnTypingStart == nil {
		log.Print("No handler for TYPING_START")
		return
	}

	var typing TypingEvent
	if err := json.Unmarshal(eventStr, &typing); err != nil {
		log.Printf("typingStart: %s", err)
		return
	}

	c.OnTypingStart(typing)
}

func (c *Client) handleEvent(eventStr []byte) {
	var event interface{}
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Print(err)
		return
	}

	eventType := event.(map[string]interface{})["t"].(string)

	// TODO: There must be a better way to directly cast the eventStr
	// to its corresponding object, avoiding double-unmarshal
	switch eventType {
	case "READY":
		c.handleReady(eventStr)
	case "MESSAGE_CREATE":
		log.Print(string(eventStr[:]))
		c.handleMessageCreate(eventStr)
	case "TYPING_START":
		c.handleTypingStart(eventStr)
	default:
		log.Printf("Ignoring %s", eventType)
		log.Printf("event dump: %s", string(eventStr[:]))
	}

}

// Get sends a GET request to the given url
func (c *Client) get(url string) (interface{}, error) {
	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.token)

	// GET the url
	log.Printf("GET %s", url)
	return do_request(req)
}

// Post sends a POST request with payload to the given url
func (c *Client) post(url string, payload interface{}) (interface{}, error) {
	pJson, _ := json.Marshal(payload)
	contentReader := bytes.NewReader(pJson)

	// Prepare request
	req, err := http.NewRequest("POST", url, contentReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	// POST the url using application/json
	log.Printf("POST %s", url)
	return do_request(req)
}

// Login initialize Discord connection by requesting a token
func (c *Client) Login(email string, password string) error {

	// Prepare POST json
	m := map[string]string{
		"email":    email,
		"password": password,
	}

	// Get token
	tokenResp, err := c.post(apiLogin, m)
	if err != nil {
		return err
	}
	c.token = tokenResp.(map[string]interface{})["token"].(string)

	// Get websocket gateway
	gatewayResp, err := c.get(apiGateway)
	if err != nil {
		return err
	}
	c.gateway = gatewayResp.(map[string]interface{})["url"].(string)

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

func (c *Client) SendMessage(channelID string, content string) error {
	_, err := c.post(
		fmt.Sprintf(apiChannels+"/%s/messages", channelID),
		map[string]string{
			"content": content,
		},
	)
	return err
}

// Run init the WebSocket connection and starts listening on it
func (c *Client) Run() {
	log.Printf("Setting up websocket to %s", c.gateway)
	conn, _, err := websocket.DefaultDialer.Dial(c.gateway, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Print("Connected")
	c.wsConn = conn

	c.doHandshake()

	for {
		_, message, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Print(err)
			break
		}
		go c.handleEvent(message)
	}
}

// Stop closes the WebSocket connection
func (c *Client) Stop() {
	log.Print("Closing connection")
	c.wsConn.Close()
}
