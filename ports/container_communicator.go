package ports

import (
	"github.com/gorilla/websocket"
)

// ContainerCommunicator defines the interface for communicating with worker containers.
type ContainerCommunicator interface {
	SendMessage(containerID string, message []byte) error
	ReceiveMessage(containerID string) error
	RemoveConnection(containerID string)
	AddConnection(containerID string, conn *websocket.Conn)
	ListConnections() []string
}
