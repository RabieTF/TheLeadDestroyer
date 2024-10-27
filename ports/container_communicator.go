package ports

type ContainerCommunicator interface {
	Send(message []byte) error
	Receive() ([]byte, error)
	HandleDisconnect() error
}
