package sockets

import (
	"encoding/json"
	"time"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions

type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, c *Client) error

const (
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
)

// SendMessageEvent is the payload sent in the
// send_message event

type UserInMessageType struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Field    string `json:"field"`
	Email    string `json:"email"`
	CodeName string `json:"codeName"`
	Address  string `json:"Address"`
}

type MessageTypeInSocket struct {
	Body           string              `json:"body"`
	ConversationId string              `json:"conversationId"`
	SeenBy         []UserInMessageType `json:"seenBy"`
	Sender         UserInMessageType   `json:"sender"`
	Receiver       UserInMessageType   `json:"receiver"`
	ID             string              `json:"_id"`
}

type SendMessageEvent struct {
	MessageTypeInSocket
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

// ReceiveMessageEvent is the payload for the "receive_message" event
type ReceiveMessageEvent struct {
	MessageTypeInSocket
}
