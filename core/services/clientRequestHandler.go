package services

import (
	"log"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket"
)

// Handler manages client requests and forwards them to the task distributor.
type ClientRequestHandler struct {
	clientWSAdapter *websocket.ContainerWebSocketAdapter
	taskDistributor *TaskDistributor
}

// NewHandler creates a new Handler instance.
func NewHandler(clientWSAdapter *websocket.ContainerWebSocketAdapter, taskDistributor *TaskDistributor) *ClientRequestHandler {
	return &ClientRequestHandler{
		clientWSAdapter: clientWSAdapter,
		taskDistributor: taskDistributor,
	}
}

// Start begins listening for client requests.
func (h *ClientRequestHandler) Start() {
	log.Println("Client request handler started")
	for {
		select {
		case hash := <-h.clientWSAdapter.HashChannel:
			// Forward the hash to the task distributor
			h.taskDistributor.TaskChannel <- hash
			log.Printf("Received hash from client: %s\n", hash)
		}
	}
}
