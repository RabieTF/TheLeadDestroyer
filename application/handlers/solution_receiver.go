package handlers

import (
	"fmt"
	"log"
	"strings"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type SolutionReceiver struct {
	containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter
	resultChannel      chan string
	distributor        *TaskDistributor
}

func NewSolutionReceiver(containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter, resultChannel chan string, distributor *TaskDistributor) *SolutionReceiver {
	return &SolutionReceiver{
		containerWSAdapter: containerWSAdapter,
		resultChannel:      resultChannel,
		distributor:        distributor,
	}
}

func (s *SolutionReceiver) Start() {
	log.Println("SolutionReceiver started")
	fmt.Println(s.distributor)
	for message := range s.containerWSAdapter.SolutionChannel {
		select {
		case s.resultChannel <- message:
			fmt.Println(message)
			hash, sol := strings.Split(message, " ")[1], strings.Split(message, " ")[2]
			fmt.Println("Solution received", hash, sol)
			s.distributor.markWorkerAvailable(s.distributor.GetWorkerFromHash(hash))
			log.Printf("Forwarded result to resultChannel: %s\n", sol)
		default:
			log.Printf("ResultChannel is full. Dropped message: %s\n", message)
		}
	}
}
