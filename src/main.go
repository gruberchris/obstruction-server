package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type GameMessage struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type PlayerJoined struct {
	Username string `json:"username"`
}

type GameBegin struct {
	PlayerOne string `json:"playerOne"`
	PlayerTwo string `json:"playerTwo"`
}

type PlayerMove struct {
	Username string `json:"username"`
	XPos     int    `json:"xpos"`
	YPos     int    `json:"ypos"`
}

type GameOver struct {
	WinningPlayer string `json:"winningPlayer"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan PlayerJoined)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// websocket route
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("Obstruction server started.")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}

	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg PlayerJoined

		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)

			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
