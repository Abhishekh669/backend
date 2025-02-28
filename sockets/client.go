package sockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

type Conversation struct {
	conversationId string
	participants   []string
	clients        map[string]*Client
}

type ConversationList map[string]*Conversation //conversationId -> coversation

type Client struct {
	UserId     string
	Connection *websocket.Conn
	Manager    *Manager

	// When any process wants to write on the client’s connection,
	// it will instead write the message to the unbuffered channel,
	//  which will block if there is any other process currently writing.
	// This makes us able to avoid any concurrency problems.

	Egress chan Event
}

type ClientList map[string]*Client //userId -> Client mapping

// func NewConversation(conn *websocket.Conn)
func (m *Manager) addConversation(conversationId string, participants []string) {
	m.Lock()
	defer m.Unlock()

	m.Conversations[conversationId] = &Conversation{
		conversationId: conversationId,
		participants:   participants,
		clients:        make(map[string]*Client),
	}
}

// todo add the user id
func NewClient(conn *websocket.Conn, manager *Manager, userId string) *Client {
	return &Client{
		UserId:     userId,
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// Graceful Close the Connection once this
		// function is done
		c.Manager.removeClient(c.UserId)
	}()

	c.Connection.SetReadLimit(4096)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := c.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.Connection.SetPongHandler(c.pongHandler)
	//loop forever  untill connection is alive

	for {

		_, payload, err := c.Connection.ReadMessage()

		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading emssage : %v", err)
			}
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			break // Breaking the connection here might be harsh xD
		}

		var payloadString string
		if err := json.Unmarshal(request.Payload, &payloadString); err != nil {
			log.Printf("error decoding payload string: %v", err)
			continue // Skip this message and continue processing others
		}

		request.Payload = json.RawMessage(payload)

		// Route the Event
		if err := c.Manager.routeEvent(request, c); err != nil {
			log.Println("Error handeling Message: ", err)
		}

	}

}

// write messages
func (c *Client) writeMessages() { //The function to handle any messages that are suppose to be sent
	//creates a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Manager.removeClient(c.UserId)
	}()

	defer close(c.Egress)

	for {
		select {
		case message, ok := <-c.Egress:
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := c.Connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				//return to close the go-routine
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				log.Println(err)
				return
			}

			// var new_data = json.RawMessage(string(data))

			// fmt.Println("Data before sending ", new_data)

			// Write a Regular text message to the connection
			if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		case <-ticker.C:
			log.Println("pinged by : ", c.UserId)
			if err := c.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return // return to break this goroutine triggeing cleanup
			}

		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong by : ", c.UserId)
	return c.Connection.SetReadDeadline(time.Now().Add(pongWait))
}
