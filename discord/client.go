package discord

import (
	"bytes"
	"encoding/json"
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
	// Handles READY
	OnReady func(ReadyEvent)
	// Handles MESSAGE_CREATE
	OnMessageReceived func(MessageEvent)

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

// Get sends a GET request to the given url
func (c *Client) Get(url string) (interface{}, error) {
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
func (c *Client) Post(url string, payload interface{}) (interface{}, error) {
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
	tokenResp, err := c.Post(apiLogin, m)
	if err != nil {
		return err
	}
	c.token = tokenResp.(map[string]interface{})["token"].(string)

	// Get websocket gateway
	gatewayResp, err := c.Get(apiGateway)
	if err != nil {
		return err
	}
	c.gateway = gatewayResp.(map[string]interface{})["url"].(string)

	return nil
}

type fileCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginFromFile call login with email and password found in the given file
func (c *Client) LoginFromFile(filename string) error {
	fileDump, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var creds = fileCredentials{}
	if err := json.Unmarshal(fileDump, &creds); err != nil {
		return err
	}

	return c.Login(creds.Email, creds.Password)
}

// Stop closes the WebSocket connection
func (c *Client) Stop() {
	log.Print("Closing connection")
	c.wsConn.Close()
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

func (c *Client) handleReady(eventStr []byte) {
	var ready ReadyEvent
	if err := json.Unmarshal(eventStr, &ready); err != nil {
		log.Printf("startKeepalive: %s", err)
		return
	}

	go func() {
		ticker := time.NewTicker(ready.Data.HeartbeatInterval * time.Millisecond)
		for range ticker.C {
			timestamp := int(time.Now().Unix())
			log.Print("Sending keepalive with timestamp %d", timestamp)
			c.wsConn.WriteJSON(map[string]int{
				"op": 1,
				"d":  timestamp,
			})
		}
	}()

	if c.OnReady == nil {
		log.Print("No handler for READY")
	} else {
		c.OnReady(ready)
	}
}

func (c *Client) handleMessageCreate(eventStr []byte) {
	if c.OnMessageReceived == nil {
		log.Print("No handler for MESSAGE_CREATE")
	} else {
		var message MessageEvent
		if err := json.Unmarshal(eventStr, &message); err != nil {
			log.Printf("messageCreate: %s", err)
		} else {
			c.OnMessageReceived(message)
		}
	}
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
		c.handleMessageCreate(eventStr)
	default:
		log.Printf("Ignoring %s", eventType)
	}

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
