package sockets

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

var (
	managerInstance *Manager
	once            sync.Once
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
)

type Manager struct {
	Clients       ClientList
	Conversations ConversationList

	//using a syncmutex here to be able ot lock state before editing clients
	//could also use channels to block

	sync.RWMutex

	//handlers are functions that are used to handle events
	Handlers map[string]EventHandler
}

func GetManager() *Manager {
	once.Do(func() {
		managerInstance = NewManager() // This ensures the manager is initialized only once
	})
	return managerInstance
}

func SendMessageHandler(event Event, c *Client) error {
	var chatevent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadMessage NewMessageEvent
	broadMessage.Sent = time.Now()
	broadMessage.Body = chatevent.Body                     // setting the `Body` field from the `MessageTypeInSocket`
	broadMessage.ConversationId = chatevent.ConversationId // setting the `ConversationId`
	broadMessage.Sender = chatevent.Sender                 // setting the `Sender`
	broadMessage.Receiver = chatevent.Receiver             // setting the `Receiver`
	broadMessage.ID = chatevent.ID                         // setting the `ID`

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage
	fmt.Println("this ist he outgoing envet : ", outgoingEvent.Type, " :: ", string(outgoingEvent.Payload))
	for userId, client := range c.Manager.Clients {
		// if userId == client.userId || userId == broadMessage.To {
		// }

		if userId == broadMessage.Receiver.ID {
			client.Egress <- outgoingEvent

		}
	}
	return nil
}

// used to initalize all the values inside the manager
func NewManager() *Manager {
	m := &Manager{
		Clients:       make(ClientList),
		Conversations: make(ConversationList),
		Handlers:      make(map[string]EventHandler),
	} //returns the pointer of the struct i.e memory address
	m.setUpEventHandlers()
	return m
}

func (m *Manager) setUpEventHandlers() {
	m.Handlers[EventSendMessage] = SendMessageHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.Handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) { //this is the method for the Manager struct and it has access to the Manager's filed
	userId := r.URL.Query().Get("userId")
	if userId == "" {
		log.Println("parameter  is missing")
		http.Error(w, "parameter is missing", http.StatusBadRequest)
		return
	}

	// if err != nil {
	// 	http.Error(w, "invalid session", http.StatusBadRequest)
	// 	return

	// }
	// fmt.Println("this is authentication in the socket : ", sessionData.Authenticated)
	log.Println("New User  connected : ", userId)

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed in connection ", err)
		http.Error(w, "failed to upgrade", http.StatusBadRequest)
		return
	}

	// create new client
	client := NewClient(conn, m, userId)

	m.addClient(client)

	m.Lock()

	for _, conversation := range m.Conversations {
		for _, participant := range conversation.participants {
			if participant == userId {
				conversation.clients[userId] = client
				break
			}
		}
	}

	go client.readMessages() //websocket connectin is only allwed to have on concurrent writter
	go client.writeMessages()

}

//add client

func (m *Manager) addClient(client *Client) {
	m.Lock()         //lock so we can manipulate one at once
	defer m.Unlock() //will execute just beforefunction is finished
	m.Clients[client.UserId] = client

}

// delete client
func (m *Manager) removeClient(userId string) {
	m.Lock()
	defer m.Unlock()

	for _, conversation := range m.Conversations {
		delete(conversation.clients, userId)
	}

	if client, ok := m.Clients[userId]; ok {
		client.Connection.Close()
		delete(m.Clients, userId)
		log.Printf("Client %s removed", userId)
	} else {
		log.Printf("Client %s not found", userId)
	}
}

// check origin
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:3000":
		return true
	default:
		return false
	}

}
