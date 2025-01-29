package websocket_adapter

import "log"

// HandleDisconnectionError logs disconnection errors.
func HandleDisconnectionError(err error) {
	if err != nil {
		log.Printf("[WARN] Disconnection error: %v\n", err)
	}
}

// HandleConnectionError logs connection errors.
func HandleConnectionError(err error) {
	if err != nil {
		log.Printf("[FATAL] Connection error: %v\n", err)
	}
}

// HandleSendError logs errors when sending messages.
func HandleSendError(containerID string, err error) {
	if err != nil {
		log.Printf("[WARN] Failed to send message to container %s: %v\n", containerID, err)
	}
}

// HandleReceiveError logs errors when receiving messages.
func HandleReceiveError(containerID string, err error) {
	if err != nil {
		log.Printf("[WARN] Failed to receive message from container %s: %v\n", containerID, err)
	}
}

// HandleUnexpectedError logs unexpected errors.
func HandleUnexpectedError(err error) {
	if err != nil {
		log.Printf("[FATAL] Unexpected error: %v\n", err)
	}
}
