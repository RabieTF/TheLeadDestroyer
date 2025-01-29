package main

import (
	"context"
	"fmt"

	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/application/handlers"
)

func main() {
	ctx := context.Background()

	containerWSAdapter := websocket_adapter.NewContainerWebSocketAdapter()

	// Initialize TaskDistributor
	swarmAdapter, err := docker.New("MD5Destroye")
	if err != nil {
		panic(err)
	}
	fmt.Println(swarmAdapter)
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
