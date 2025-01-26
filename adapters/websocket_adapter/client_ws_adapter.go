package websocket_adapter

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// ClientWebSocketAdapter implements the ClientCommunicator interface for a single WebSocket client.
type ClientWebSocketAdapter struct {
	conn *websocket.Conn // Single WebSocket connection
}

// NewClientWebSocketAdapter creates a new instance of ClientWebSocketAdapter with a WebSocket connection.
func NewClientWebSocketAdapter(conn *websocket.Conn) *ClientWebSocketAdapter {
	log.Println("Client connected")
	return &ClientWebSocketAdapter{
		conn: conn,
	}
}

// Send sends a message to the connected client.
func (adapter *ClientWebSocketAdapter) Send(message []byte) error {
	if adapter.conn == nil {
		return fmt.Errorf("no client connected")
	}
	err := adapter.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return err
	}
	return nil
}

// Receive receives a message from the connected client.
func (adapter *ClientWebSocketAdapter) Receive() ([]byte, error) {
	if adapter.conn == nil {
		return nil, fmt.Errorf("no client connected")
	}
	_, message, err := adapter.conn.ReadMessage()
	if err != nil {
		log.Printf("Error receiving message: %v\n", err)
		return nil, err
	}
	return message, nil
}

// HandleDisconnect gracefully handles the disconnection of the client.
func (adapter *ClientWebSocketAdapter) HandleDisconnect() error {
	if adapter.conn == nil {
		return fmt.Errorf("no client connected")
	}
	err := adapter.conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v\n", err)
		return err
	}
	adapter.conn = nil
	log.Println("Client disconnected")
	return nil
}
