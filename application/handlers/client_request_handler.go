package handlers

import (
	"log"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type ClientRequestHandler struct {
	clientWSAdapter *websocket_adapter.ClientWebSocketAdapter
	taskDistributor *TaskDistributor
	resultChannel   chan string // Channel to receive results from SolutionReceiver
}

// NewClientRequestHandler creates a new ClientRequestHandler instance.
func NewClientRequestHandler(clientWSAdapter *websocket_adapter.ClientWebSocketAdapter, taskDistributor *TaskDistributor, resultChannel chan string) *ClientRequestHandler {
	return &ClientRequestHandler{
		clientWSAdapter: clientWSAdapter,
		taskDistributor: taskDistributor,
		resultChannel:   resultChannel,
	}
}

// Start begins handling client requests and results.
func (h *ClientRequestHandler) Start() {
	log.Println("ClientRequestHandler started")

	// Start handling client requests
	go h.handleClientRequests()

	// Start forwarding results to the client
	go h.forwardResultsToClient()
}

// handleClientRequests listens for messages from the client and forwards them to the TaskDistributor's TaskChannel.
func (h *ClientRequestHandler) handleClientRequests() {
	for {
		// Receive message from the client
		message, err := h.clientWSAdapter.Receive()
		if err != nil {
			log.Printf("Error receiving message: %v\n", err)
			if disconnectErr := h.clientWSAdapter.HandleDisconnect(); disconnectErr != nil {
				log.Printf("Error handling disconnect: %v\n", disconnectErr)
			}
			break
		}

		hash := string(message)
		log.Printf("Received hash: %s\n", hash)

		// Forward the hash to the TaskDistributor's TaskChannel
		select {
		case h.taskDistributor.TaskChannel <- hash:
			log.Printf("Hash %s sent to TaskDistributor\n", hash)
		default:
			log.Printf("TaskChannel is full, unable to send hash: %s\n", hash)
		}
	}
}

// forwardResultsToClient listens for results from the resultChannel and sends them back to the client.
func (h *ClientRequestHandler) forwardResultsToClient() {
	for result := range h.resultChannel {
		log.Printf("Forwarding result to client: %s\n", result)

		// Send the result to the client
		err := h.clientWSAdapter.Send([]byte(result))
		if err != nil {
			log.Printf("Error sending result to client: %v\n", err)
			h.clientWSAdapter.HandleDisconnect()
			return
		}
	}
}
