package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type TaskDistributor struct {
	TaskChannel        chan string
	containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter
	swarmAdapter       *docker.Adapter
	mu                 sync.Mutex
	activeWorkers      map[string]bool // Tracks active worker availability
	minReplicas        int
	maxReplicas        int
	threshold          int // Tasks per worker before scaling up
}

// NewDistributor creates a new Distributor instance.
func NewDistributor(containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter, swarmAdapter *docker.Adapter, minReplicas, maxReplicas, threshold int) *TaskDistributor {
	return &TaskDistributor{
		TaskChannel:        make(chan string, 100),
		containerWSAdapter: containerWSAdapter,
		swarmAdapter:       swarmAdapter,
		activeWorkers:      make(map[string]bool),
		minReplicas:        minReplicas,
		maxReplicas:        maxReplicas,
		threshold:          threshold,
	}
}

// Start begins distributing tasks and dynamically scaling workers.
func (d *TaskDistributor) Start(ctx context.Context) {
	log.Println("Task distributor started")
	ticker := time.NewTicker(5 * time.Second) // Periodic scaling check
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Task distributor shutting down")
			return

		case hash := <-d.TaskChannel:
			workerID, err := d.getAvailableWorker()
			if err != nil {
				log.Printf("No available worker for hash: %s. Retrying later.\n", hash)
				go d.retryTask(hash)
				continue
			}

			if err := d.assignTaskToWorker(workerID, hash); err != nil {
				log.Printf("Failed to assign task to worker %s: %v. Retrying task.\n", workerID, err)
				go d.retryTask(hash)
			}

		case <-ticker.C:
			d.manageScaling(ctx)
		}
	}
}

// manageScaling scales workers up or down based on the number of tasks in the queue.
func (d *TaskDistributor) manageScaling(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()

	queueSize := len(d.TaskChannel)
	workerCount := len(d.activeWorkers)
	desiredReplicas := d.calculateReplicas(queueSize)

	if desiredReplicas > workerCount {
		log.Printf("Scaling up to %d replicas (current: %d, queue: %d tasks)\n", desiredReplicas, workerCount, queueSize)
		err := d.swarmAdapter.ScaleService(ctx, uint64(desiredReplicas))
		if err != nil {
			return
		}
		d.refreshWorkers(ctx)
	} else if desiredReplicas < workerCount && workerCount > d.minReplicas {
		log.Printf("Scaling down to %d replicas (current: %d, queue: %d tasks)\n", desiredReplicas, workerCount, queueSize)
		err := d.swarmAdapter.ScaleService(ctx, uint64(desiredReplicas))
		if err != nil {
			return
		}
		d.refreshWorkers(ctx)
	}
}

// calculateReplicas determines the number of replicas based on the queue size.
func (d *TaskDistributor) calculateReplicas(queueSize int) int {
	replicas := int(math.Ceil(float64(queueSize) / float64(d.threshold)))
	if replicas < d.minReplicas {
		return d.minReplicas
	}
	if replicas > d.maxReplicas {
		return d.maxReplicas
	}
	return replicas
}

// refreshWorkers updates the active worker list after scaling.
func (d *TaskDistributor) refreshWorkers(ctx context.Context) {
	workerIPs, err := d.swarmAdapter.GetContainerIPs(ctx)
	if err != nil {
		log.Printf("Failed to refresh workers: %v\n", err)
		return
	}

	d.activeWorkers = make(map[string]bool)
	for _, ip := range workerIPs {
		d.activeWorkers[ip] = true // Mark all as available
	}

	log.Printf("Active workers refreshed: %v\n", d.activeWorkers)
}

// getAvailableWorker retrieves an available worker.
func (d *TaskDistributor) getAvailableWorker() (string, error) {
	for workerID, isAvailable := range d.activeWorkers {
		if isAvailable {
			d.activeWorkers[workerID] = false // Mark as busy
			return workerID, nil
		}
	}
	return "", errors.New("no available workers")
}

// assignTaskToWorker assigns a task to a worker.
func (d *TaskDistributor) assignTaskToWorker(workerID, hash string) error {
	// Define the 4-character range for brute force
	begin := "0000"
	end := "ZZZZ"

	// Construct the search message
	message := fmt.Sprintf("search %s %s %s", hash, begin, end)

	// Send the message to the worker
	if err := d.containerWSAdapter.SendMessage(workerID, []byte(message)); err != nil {
		d.markWorkerAvailable(workerID) // Mark the worker as available again on failure
		return err
	}

	log.Printf("Assigned hash %s to worker %s\n", hash, workerID)
	return nil
}

// retryTask re-queues a task for retrying.
func (d *TaskDistributor) retryTask(hash string) {
	time.Sleep(2 * time.Second) // Optional delay before retry
	d.TaskChannel <- hash
}

// markWorkerAvailable marks a worker as available.
func (d *TaskDistributor) markWorkerAvailable(workerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.activeWorkers[workerID] = true
}
