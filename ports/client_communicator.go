package ports

type ClientCommunicator interface {
	Send(message []byte) error
	Receive() ([]byte, error)
	HandleDisconnect() error
}
