package handlers

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/docker"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type TaskDistributor struct {
	TaskChannel        chan string
	currentQueue       *list.List
	containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter
	swarmAdapter       *docker.Adapter
	mu                 sync.Mutex
	activeWorkers      map[string]string // Tracks active worker availability "" and unavailability "hash"
	minReplicas        int
	maxReplicas        int
	threshold          int // Tasks per worker before scaling up
}

// NewDistributor creates a new Distributor instance.
func NewDistributor(containerWSAdapter *websocket_adapter.ContainerWebSocketAdapter, swarmAdapter *docker.Adapter, minReplicas, maxReplicas, threshold int) *TaskDistributor {
	return &TaskDistributor{
		TaskChannel:        make(chan string, 100),
		currentQueue:       list.New(),
		containerWSAdapter: containerWSAdapter,
		swarmAdapter:       swarmAdapter,
		activeWorkers:      make(map[string]string),
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
			d.currentQueue.PushBack(hash)
			workerID, err := d.getAvailableWorker()
			if err != nil {
				log.Printf("No available worker for hash: %s. Retrying later.\n", hash)
				go d.retryTask(hash)
				continue
			}

			hash = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(hash, " ", ""), "\n", ""), "\t", "")

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

	queueSize := d.currentQueue.Len()
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
	activeConnections := d.containerWSAdapter.ListConnections()
	fmt.Printf("Active connections: %d\n", len(activeConnections))
	d.activeWorkers = make(map[string]string)
	for _, id := range activeConnections {
		d.activeWorkers[id] = "" // Mark all as available
	}

	log.Printf("Active workers refreshed: %v\n", d.activeWorkers)
}

// getAvailableWorker retrieves an available worker.
func (d *TaskDistributor) getAvailableWorker() (string, error) {
	for workerID, task := range d.activeWorkers {
		if task == "" {
			return workerID, nil
		}
	}
	return "", errors.New("no available workers")
}

// assignTaskToWorker assigns a task to a worker.
func (d *TaskDistributor) assignTaskToWorker(workerID, hash string) error {
	// Define the 4-character range for brute force
	begin := "0"
	end := "ZZZZ"

	// Construct the search message
	message := fmt.Sprintf("search %s %s %s", hash, begin, end)

	d.activeWorkers[workerID] = hash

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
	fmt.Println("Marking worker as available: ", workerID)
	defer d.mu.Unlock()
	d.activeWorkers[workerID] = ""
}

func (d *TaskDistributor) GetWorkerFromHash(hash string) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, v := range d.activeWorkers {
		if v == hash {
			return v
		}
	}
	return ""
}

func (d *TaskDistributor) RemoveHashFromQueue(hash string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.activeWorkers, hash)
}

type ContainerInfo struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Status  string `json:"status"`
	Hash    string `json:"hash"`
}

func (d *TaskDistributor) GetContainersInfo() (*[]ContainerInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var containers []ContainerInfo

	for workerID, assignedHash := range d.activeWorkers {
		status := "inactif"
		if assignedHash != "" {
			status = "actif"
		}

		container := ContainerInfo{
			ID:      workerID,
			GroupID: "default",
			Status:  status,
			Hash:    assignedHash,
		}

		containers = append(containers, container)
	}

	if len(containers) == 0 {
		return nil, errors.New("no active containers found")
	}

	return &containers, nil
}
