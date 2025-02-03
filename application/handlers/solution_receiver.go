package handlers

import (
	"log"
	"strings"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type SolutionReceiver struct {
	containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter
	resultChannel      chan string
	distributor        *TaskDistributor
}

func NewSolutionReceiver(containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter, resultChannel chan string) *SolutionReceiver {
	return &SolutionReceiver{
		containerWSAdapter: containerWSAdapter,
		resultChannel:      resultChannel,
	}
}

func (s *SolutionReceiver) Start() {
	log.Println("SolutionReceiver started")
	for message := range s.containerWSAdapter.SolutionChannel {
		select {
		case s.resultChannel <- message:
			id, msg := strings.Split(message, ":")[0], strings.Split(message, ":")[1]
			s.distributor.markWorkerAvailable(id)
			log.Printf("Forwarded result to resultChannel: %s\n", msg)
		default:
			log.Printf("ResultChannel is full. Dropped message: %s\n", message)
		}
	}
}
