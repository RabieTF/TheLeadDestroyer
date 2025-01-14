package ports

// ClientCommunicator defines the interface for communicating with clients.
type ClientCommunicator interface {
	// Send sends a message to a specific client.
	Send(clientID string, message []byte) error

	// Receive receives a message from a specific client.
	Receive(clientID string) ([]byte, error)

	// HandleDisconnect disconnects a specific client.
	HandleDisconnect(clientID string) error

	// AddClient adds a new client connection.
	AddClient(clientID string, conn interface{}) error

	// RemoveClient removes a client connection.
	RemoveClient(clientID string) error
}
