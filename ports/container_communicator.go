package ports

// ContainerCommunicator defines the interface for communicating with worker containers.
type ContainerCommunicator interface {
	// Send sends a message to a specific container.
	Send(containerID string, message []byte) error

	// Receive receives a message from a specific container.
	Receive(containerID string) ([]byte, error)

	// HandleDisconnect disconnects a specific container.
	HandleDisconnect(containerID string) error

	// AddConnection adds a new container connection.
	AddConnection(containerID, url string) error

	// ListConnections returns the IDs of all connected containers.
	ListConnections() []string
}
