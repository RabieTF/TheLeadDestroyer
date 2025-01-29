package websocket_adapter

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ContainerWebSocketAdapter struct {
	connections     map[string]*websocket.Conn // Maps container IDs to WebSocket connections
	mux             sync.Mutex
	SolutionChannel chan string // Channel for forwarding results to SolutionReceiver
}

func NewContainerWebSocketAdapter() *ContainerWebSocketAdapter {
	return &ContainerWebSocketAdapter{
		connections:     make(map[string]*websocket.Conn),
		SolutionChannel: make(chan string, 100),
	}
}

func (c *ContainerWebSocketAdapter) AddConnection(containerID string, conn *websocket.Conn) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.connections[containerID] = conn
	log.Printf("Container %s connected.\n", containerID)
}

func (c *ContainerWebSocketAdapter) RemoveConnection(containerID string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if conn, exists := c.connections[containerID]; exists {
		conn.Close()
		delete(c.connections, containerID)
		log.Printf("Container %s disconnected.\n", containerID)
	}
}

func (c *ContainerWebSocketAdapter) SendMessage(containerID string, message []byte) error {
	c.mux.Lock()
	conn, ok := c.connections[containerID]
	c.mux.Unlock()
	if !ok {
		return fmt.Errorf("container %s not connected", containerID)
	}

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Failed to send message to container %s: %v\n", containerID, err)
	}
	return err
}

func (c *ContainerWebSocketAdapter) ReceiveMessage(containerID string) error {
	conn, ok := c.connections[containerID]
	if !ok {
		return fmt.Errorf("container %s not connected", containerID)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Error receiving message from container %s: %v\n", containerID, err)
		return err
	}

	msg := string(message)
	c.SolutionChannel <- msg
	log.Printf("Message received from container %s: %s\n", containerID, msg)
	return nil
}

func (c *ContainerWebSocketAdapter) ListConnections() []string {
	c.mux.Lock()
	defer c.mux.Unlock()

	ids := make([]string, 0, len(c.connections))
	for id := range c.connections {
		ids = append(ids, id)
	}
	return ids
}
