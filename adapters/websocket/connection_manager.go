package websocket

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// ConnectionManager manages active WebSocket connections.
type ConnectionManager struct {
	clients map[*websocket.Conn]bool
	mutex   sync.Mutex
	rdb     *redis.Client // Redis client
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager(rdb *redis.Client) *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[*websocket.Conn]bool),
		rdb:     rdb,
	}
}

// getHash retrieves a cracked hash from Redis.
func (cm *ConnectionManager) getHash(hash string) (string, error) {
	if cm.rdb == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	ctx := context.Background()
	password, err := cm.rdb.Get(ctx, "hash:"+hash).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("Hash not found")
	} else if err != nil {
		return "", err
	}
	return password, nil
}

// storeHash stores a cracked hash in Redis.
func (cm *ConnectionManager) storeHash(hash, password string) error {
	if cm.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	ctx := context.Background()
	return cm.rdb.Set(ctx, "hash:"+hash, password, 0).Err()
}

// AddClient adds a new connection to the manager.
func (cm *ConnectionManager) AddClient(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.clients[conn] = true
	log.Println("New client connected")
}

// RemoveClient removes a connection from the manager.
func (cm *ConnectionManager) RemoveClient(conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.clients, conn)
	log.Println("Client disconnected")
}

// Broadcast sends a message to all connected clients.
func (cm *ConnectionManager) Broadcast(message string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	for conn := range cm.clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error broadcasting message:", err)
			conn.Close()
			delete(cm.clients, conn)
		}
	}
}

// SendMessage sends a message to a specific client.
func (cm *ConnectionManager) SendMessage(conn *websocket.Conn, message string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	if _, ok := cm.clients[conn]; !ok {
		return fmt.Errorf("client not found")
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
