package ports

type ContainerCommunicator interface {
	SendMessage(message []byte) error
	ReceiveMessage() ([]byte, error)
	HandleClientDisconnect() error
}