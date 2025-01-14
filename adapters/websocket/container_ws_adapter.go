package websocket

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ContainerWebSocketAdapter struct {
	connections     map[string]*websocket.Conn // Maps container IDs to WebSocket connections
	mux             sync.Mutex
	SolutionChannel chan string
}

func NewContainerWebSocketAdapter() *ContainerWebSocketAdapter {
	return &ContainerWebSocketAdapter{connections: make(map[string]*websocket.Conn)}
}

func (c *ContainerWebSocketAdapter) connect(url string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		HandleConnectionError(err) // Handle connection error
		return nil, fmt.Errorf("failed to connect to %s: %w", url, err)
	}
	return conn, nil
}

func (c *ContainerWebSocketAdapter) AddConnection(containerID, url string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, exists := c.connections[containerID]; exists {
		return fmt.Errorf("connection for container %s already exists", containerID)
	}

	conn, err := c.connect(url)
	if err != nil {
		return err
	}

	c.connections[containerID] = conn
	log.Printf("Container %s connected to %s\n", containerID, url)
	return nil
}

func (c *ContainerWebSocketAdapter) SendMessage(containerID string, message []byte) error {
	c.mux.Lock()
	conn, ok := c.connections[containerID]
	c.mux.Unlock()
	if !ok {
		return fmt.Errorf("container %s not connected", containerID)
	}
	err := conn.WriteMessage(websocket.TextMessage, message)
	HandleSendError(containerID, err)
	return err
}

func (c *ContainerWebSocketAdapter) ReceiveMessage(containerID string) error {
	conn, ok := c.connections[containerID]
	if !ok {
		return fmt.Errorf("container %s not connected", containerID)
	}
	_, message, err := conn.ReadMessage()
	HandleReceiveError(containerID, err)

	c.SolutionChannel <- string(message)
	return err
}

func (c *ContainerWebSocketAdapter) HandleDisconnect(containerID string) error {
	c.mux.Lock()
	conn, ok := c.connections[containerID]
	if ok {
		delete(c.connections, containerID)
	}
	c.mux.Unlock()

	if conn == nil {
		return nil
	}

	if err := conn.Close(); err != nil {
		HandleDisconnectionError(err)
		return err
	}

	log.Printf("Container %s disconnected\n", containerID)
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
