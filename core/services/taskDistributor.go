package services

import (
	"log"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket"
)

// TaskDistributor manages the distribution of tasks to workers.
type TaskDistributor struct {
	TaskChannel        chan string
	containerWSAdapter *websocket.ContainerWebSocketAdapter
	swarmAdapter       *docker.Adapter
}

// NewDistributor creates a new Distributor instance.
func NewDistributor(containerWSAdapter *websocket.ContainerWebSocketAdapter, swarmAdapter *docker.Adapter) *TaskDistributor {
	return &TaskDistributor{
		TaskChannel:        make(chan string, 100),
		containerWSAdapter: containerWSAdapter,
		swarmAdapter:       swarmAdapter,
	}
}

// Start begins distributing tasks to workers.
func (d *TaskDistributor) Start() {
	log.Println("Task distributor started")
	for {
		select {
		case hash := <-d.TaskChannel:
			// Send the hash to an available worker
			d.containerWSAdapter.SendHash(hash)
			log.Printf("Distributed hash to worker: %s\n", hash)
		}
	}
}
