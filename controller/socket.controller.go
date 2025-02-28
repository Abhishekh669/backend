package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ID   string
	Conn *websocket.Conn
}

var clients = make(map[string]*Client)

func WebSocketHanlder(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("error upgrading connection", err)
		return
	}

	clientID := uuid.New().String()

	client := &Client{
		ID:   clientID,
		Conn: conn,
	}
	clients[clientID] = client

	defer conn.Close()
	fmt.Println("Client connected")

	for {
		messageType, message, err := conn.ReadMessage()

		if err != nil {
			fmt.Println("error reading message : ", err)

		}

		fmt.Println("message received : ", string(message), "ID : ", clientID)

		time.Sleep(3 * time.Second)

		if err := conn.WriteMessage(messageType, message); err != nil {
			fmt.Println("error writing messages : ", err)
			break
		}
	}

	delete(clients, clientID)
}

func printAllClients() {
	fmt.Println("All connected clients:")
	for clientID, client := range clients {
		fmt.Printf("Client ID: %s, Connection: %v\n", clientID, client.Conn)
	}
}
