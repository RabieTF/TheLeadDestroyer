package main

import (
	"context"

	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/application/handlers"
)

func main() {
	ctx := context.Background()

	containerWSAdapter := websocket_adapter.NewContainerWebSocketAdapter()

	// Initialize TaskDistributor
	swarmAdapter, _ := docker.New("my-service")
	taskDistributor := handlers.NewDistributor(containerWSAdapter, swarmAdapter, 2, 10, 5)
	go taskDistributor.Start(ctx)

	// Initialize SolutionReceiver
	resultChannel := make(chan string, 100)
	solutionReceiver := handlers.NewSolutionReceiver(containerWSAdapter, resultChannel)
	go solutionReceiver.Start()

	// Initialize ConnectionFactory
	connectionFactory := handlers.NewConnectionFactory(containerWSAdapter, taskDistributor, resultChannel)

	// Start the WebSocket server
	connectionFactory.StartServer("3000")
}
