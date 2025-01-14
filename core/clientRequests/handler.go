package clientRequests

import (
	"log"
	"theleaddestroyer/adapters/websocket"
	"theleaddestroyer/core/taskDistribution"
)

// Handler manages client requests and forwards them to the task distributor.
type Handler struct {
	clientWSAdapter *websocket.ClientWSAdapter
	taskDistributor *taskDistribution.Distributor
}

// NewHandler creates a new Handler instance.
func NewHandler(clientWSAdapter *websocket.ClientWSAdapter, taskDistributor *taskDistribution.Distributor) *Handler {
	return &Handler{
		clientWSAdapter: clientWSAdapter,
		taskDistributor: taskDistributor,
	}
}

// Start begins listening for client requests.
func (h *Handler) Start() {
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
