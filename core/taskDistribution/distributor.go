package taskDistribution

import (
	"log"
	"theleaddestroyer/adapters/docker"
	"theleaddestroyer/adapters/websocket"
)

// Distributor manages the distribution of tasks to workers.
type Distributor struct {
	TaskChannel        chan string
	containerWSAdapter *websocket.ContainerWSAdapter
	swarmAdapter       *docker.SwarmAdapter
}

// NewDistributor creates a new Distributor instance.
func NewDistributor(containerWSAdapter *websocket.ContainerWSAdapter, swarmAdapter *docker.SwarmAdapter) *Distributor {
	return &Distributor{
		TaskChannel:        make(chan string, 100),
		containerWSAdapter: containerWSAdapter,
		swarmAdapter:       swarmAdapter,
	}
}

// Start begins distributing tasks to workers.
func (d *Distributor) Start() {
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
