package handlers

import (
	"log"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type SolutionReceiver struct {
	containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter
	resultChannel      chan string
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
			log.Printf("Forwarded result to resultChannel: %s\n", message)
		default:
			log.Printf("ResultChannel is full. Dropped message: %s\n", message)
		}
	}
}
