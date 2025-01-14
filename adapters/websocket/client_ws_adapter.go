package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// HandleWebSocket handles a WebSocket connection, receives messages, and broadcasts them.
func HandleWebSocket(cm *ConnectionManager, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins (for testing purposes)
		},
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Add the client to the connection manager
	cm.AddClient(conn)
	defer cm.RemoveClient(conn)

	log.Println("New client connected")

	// Listen for messages from clients
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v\n", err)
			} else {
				log.Println("Client disconnected")
			}
			break
		}

		hash := string(message)
		log.Printf("Received hash: %s\n", hash)

		// Check if the hash is already cracked
		password, err := cm.getHash(hash)
		if err != nil {
			log.Printf("Error checking hash in Redis: %v\n", err)
		} else {
			// Hash already cracked, send the result to the client
			log.Printf("Hash already cracked: %s -> %s\n", hash, password)
			if err := cm.SendMessage(conn, fmt.Sprintf("found %s %s", hash, password)); err != nil {
				log.Printf("Failed to send message to client: %v\n", err)
			}
			continue
		}

		// If not found, broadcast the hash to workers
		log.Printf("Broadcasting hash: %s\n", hash)
		cm.Broadcast(hash)

		// TODO :Simulate cracking the hash (replace with actual logic)
		// example, store the result in Redis
		if hash == "cad77c7dffc10fcacc77ff0690f2897a" {
			password := "pina"
			log.Printf("Simulating hash crack: %s -> %s\n", hash, password)
			if err := cm.storeHash(hash, password); err != nil {
				log.Printf("Failed to store hash in Redis: %v\n", err)
			} else {
				if err := cm.SendMessage(conn, fmt.Sprintf("found %s %s", hash, password)); err != nil {
					log.Printf("Failed to send message to client: %v\n", err)
				}
			}
		}
	}

	log.Println("Client connection closed")
}
