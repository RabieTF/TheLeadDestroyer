package ports

// ClientCommunicator defines the interface for communicating with clients.
type ClientCommunicator interface {
	// Send sends a message to the connected client.
	Send(message []byte) error

	// Receive receives a message from the connected client.
	Receive() ([]byte, error)

	// HandleDisconnect gracefully handles the disconnection of the client.
	HandleDisconnect() error
}
