package discord

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

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

type discordClient struct {
	wsConn  *websocket.Conn
	gateway string
	token   string
}

func NewClient() *discordClient {
	c := discordClient{}
	c.setupHandlers()
	return c
}

// func (c *Client) setupHandlers

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

	// Init handlers map
	c.Handlers = make(map[string]func(Event))

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

// AddHandler stores a function to be called asynchronously upon receiving the specified event
func (c *Client) AddHandler(event string, handler func(Event)) {
	log.Printf("Adding handler for %s event", event)
	c.Handlers[event] = handler
}

func (c *Client) doHandshake() {
	log.Print("Sending handshake")
	c.wsConn.WriteJSON(Event{
		OpCode: 2,
		Data: map[string]interface{}{
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

// Keepalive
// c.wsConn.WriteJSON(map[string]int{
// 	"op": 1,
// 	"d":  int(time.Now().Unix()),
// })

func (c *Client) handleEvent(eventStr []byte) {
	var event = Event{}
	if err := json.Unmarshal(eventStr, &event); err != nil {
		log.Print(err)
		return
	}

	log.Printf("Event %s received", event.Type)

	handler, ok := c.Handlers[event.Type]
	if ok {
		log.Printf("Executing handler for %s", event.Type)
		go handler(event)
	} else {
		log.Print("No handler found, ignoring")
	}
}

func (c *Client) newMessage() {

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
		c.handleEvent(message)
	}
}
