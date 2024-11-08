package websocket

import "log"

// this file is to handle the websocket errors

func HandleDisconnectionError(err error) {
	if err != nil {
		log.SetPrefix("WARN")
		log.Printf("Disconnection error: %v\n", err)
	}
}

func HandleConnectionError(err error) {
	if err != nil {
		log.SetPrefix("FATAL")
		log.Fatalf("Connection error: %v\n", err)
	}
}

func HandleSendError(containerID string, err error) {
	if err != nil {
		log.SetPrefix("WARN")
		log.Printf("Failed to send message to container %s: %v\n", containerID, err)
	}
}

func HandleReceiveError(containerID string, err error) {
	if err != nil {
		log.SetPrefix("WARN")
		log.Printf("Failed to receive message from container %s: %v\n", containerID, err)
	}
}
func HandleUnexpectedError(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
}
