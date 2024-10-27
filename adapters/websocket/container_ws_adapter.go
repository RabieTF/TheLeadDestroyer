package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type ContainerWebSocketAdapter struct {
	connections map[string]*websocket.Conn // Maps container IDs to WebSocket connections
	mux         sync.Mutex
}

func NewContainerWebSocketAdapter() *ContainerWebSocketAdapter {
	return &ContainerWebSocketAdapter{connections: make(map[string]*websocket.Conn)}
}

func (c *ContainerWebSocketAdapter) AddConnection(containerID string, conn *websocket.Conn) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.connections[containerID] = conn
}

func (c *ContainerWebSocketAdapter) SendMessage(containerID string, message []byte) error {
	c.mux.Lock()
	conn, ok := c.connections[containerID]
	c.mux.Unlock()
	if !ok {
		return fmt.Errorf("container %s not connected", containerID)
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}

func (c *ContainerWebSocketAdapter) ReceiveMessage(containerID string) ([]byte, error) {
	conn, ok := c.connections[containerID]
	if !ok {
		return nil, fmt.Errorf("container %s not connected", containerID)
	}
	_, message, err := conn.ReadMessage()
	return message, err
}

func (c *ContainerWebSocketAdapter) HandleDisconnect(containerID string) error {
	c.mux.Lock()
	conn, ok := c.connections[containerID]
	if ok {
		delete(c.connections, containerID)
	}
	c.mux.Unlock()
	if conn != nil {
		return conn.Close()
	}
	return nil
}
