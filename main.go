package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/application/handlers"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	minReplicas, err := strconv.Atoi(os.Getenv("MIN_REPLICAS"))
	maxReplicas, err := strconv.Atoi(os.Getenv("MAX_REPLICAS"))
	threshold, err := strconv.Atoi(os.Getenv("THRESHOLD"))

	if err != nil {
		log.Fatal("Please make sure env variables are integers.")
	}

	containerWSAdapter := websocket_adapter.NewContainerWebSocketAdapter()

	// Initialize TaskDistributor
	swarmAdapter, err := docker.New("md5onelettertest")
	if err != nil {
		panic(err)
	}
	taskDistributor := handlers.NewDistributor(containerWSAdapter, swarmAdapter, minReplicas, maxReplicas, threshold)
	go taskDistributor.Start(ctx)

	// Initialize SolutionReceiver
	resultChannel := make(chan string, 100)
	solutionReceiver := handlers.NewSolutionReceiver(containerWSAdapter, resultChannel, taskDistributor)
	go solutionReceiver.Start()

	// Initialize ConnectionFactory
	connectionFactory := handlers.NewConnectionFactory(containerWSAdapter, taskDistributor, resultChannel)

	// Start the WebSocket server
	connectionFactory.StartServer("8080")
}
