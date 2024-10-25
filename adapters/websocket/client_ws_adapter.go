package websocket // websocket adapter

import (
	"github.com/gorilla/websocket"
	"sync"
)

type ClientWebSocketAdapter struct {
	conn *websocket.Conn
	mux  sync.Mutex
}

func NewClientWebSocketAdapter(conn *websocket.Conn) *ClientWebSocketAdapter {
	return &ClientWebSocketAdapter{conn: conn}
}
