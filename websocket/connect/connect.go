package connect

import (
	"log"
	"net/http"
	"sync"

	"backend/concurrency/websocket/publisher"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(req *http.Request) bool {
		return true // Allow all origins
	},
}

// Global list of Clients (WebSocket connections)
var Clients = make(map[*websocket.Conn]bool)
var Mutex = sync.Mutex{}

// Broadcast channel to send messages to all clients
var broadcast = make(chan []byte)

func HandleWebSocket(writer http.ResponseWriter, req *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, readErr := upgrader.Upgrade(writer, req, nil)
	if readErr != nil {
		log.Println("Error upgrading connection:", readErr)
		return
	}

	// Add the new client to the list of clients
	Mutex.Lock()
	Clients[conn] = true
	Mutex.Unlock()

	log.Println("New connection established, client connected")

	go handleConnection(conn)
}

func handleConnection(conn *websocket.Conn) {
	defer func() {
		// Remove the client from the list of clients when the connection is closed
		Mutex.Lock()
		delete(Clients, conn)
		Mutex.Unlock()
		conn.Close()
	}()

	for {
		// Read message from client
		var message []byte
		var err error

		_, message, err = conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		log.Printf("Received message: %s", message)

		// Echo the message back to the client
		broadcast <- message
	}
}

func PublishMessage() {
	for {
		message := <-broadcast

		// // Lock the clients map to ensure thread safety
		// mutex.Lock()

		// // Broadcast the message to all clients
		// for client := range clients {
		// 	var writeErr error
		// 	if writeErr = client.WriteMessage(websocket.TextMessage, message); writeErr != nil {
		// 		log.Println("Error writing message:", writeErr)
		// 		break
		// 	}
		// }

		// mutex.Unlock()

		publisher.Send(message)
	}
}
