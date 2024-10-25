package ports

type ClientCommunicator interface {
	SendMessage(message []byte) error
	ReceiveMessage() ([]byte, error)
	HandleClientDisconnect() error
}