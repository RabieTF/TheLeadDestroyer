package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	websocketAdapter "www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket_adapter"
)

type ConnectionFactory struct {
	containerAdapter *websocketAdapter.ContainerWebSocketAdapter
	taskDistributor  *TaskDistributor
	resultChannel    chan string
}

// NewConnectionFactory initializes a new ConnectionFactory.
func NewConnectionFactory(
	containerAdapter *websocketAdapter.ContainerWebSocketAdapter,
	taskDistributor *TaskDistributor,
	resultChannel chan string,
) *ConnectionFactory {
	return &ConnectionFactory{
		containerAdapter: containerAdapter,
		taskDistributor:  taskDistributor,
		resultChannel:    resultChannel,
	}
}

// StartServer starts the WebSocket server and handles routing connections.
func (cf *ConnectionFactory) StartServer(port string) {
	http.HandleFunc("/ws", cf.HandleConnection)
	http.HandleFunc("/status", cf.handleContainersInfo)

	log.Printf("WebSocket and HTTP status server starting on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start WebSocket server: %v\n", err)
	}
}

// HandleConnection handles incoming WebSocket connections and determines their type.
func (cf *ConnectionFactory) HandleConnection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity
		},
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v\n", err)
		return
	}

	// Read the first message to determine the type of connection
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read initial message: %v\n", err)
		conn.Close()
		return
	}

	msg := strings.TrimSpace(string(message)) // Trim newlines and spaces
	log.Printf("Connection type identified: %s\n", msg)

	// Route the connection based on type
	switch msg {
	case "client":
		cf.handleClientConnection(conn)

	case "slave":
		cf.handleSlaveConnection(conn)

	default:
		log.Printf("Unknown connection type: %s. Closing connection.\n", msg)
		conn.Close()
	}
}

// handleClientConnection initializes a client connection and starts the handler.
func (cf *ConnectionFactory) handleClientConnection(conn *websocket.Conn) {
	log.Println("Initializing client connection")

	clientAdapter := websocketAdapter.NewClientWebSocketAdapter(conn)
	clientHandler := NewClientRequestHandler(clientAdapter, cf.taskDistributor, cf.resultChannel)

	go clientHandler.Start()
}

// handleSlaveConnection initializes a slave connection and listens for messages.
func (cf *ConnectionFactory) handleSlaveConnection(conn *websocket.Conn) {
	log.Println("Registering slave connection")

	slaveID := uuid.New().String()
	cf.containerAdapter.AddConnection(slaveID, conn)

	// Listen for messages from the slave
	go func() {
		defer cf.containerAdapter.RemoveConnection(slaveID)
		for {
			err := cf.containerAdapter.ReceiveMessage(slaveID)
			if err != nil {
				log.Printf("Error with slave %s: %v\n", slaveID, err)
				break
			}
		}
	}()
}

func (cf *ConnectionFactory) handleContainersInfo(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	containers, err := cf.taskDistributor.GetContainersInfo()
	if err != nil {
		log.Printf("Error getting containers info: %v\n", err)
		http.Error(w, "Failed to fetch containers", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(containers)
	if err != nil {
		log.Printf("Error marshalling containers info: %v\n", err)
		http.Error(w, "Failed to process data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
