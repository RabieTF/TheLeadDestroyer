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
func (c *ClientWebSocketAdapter) Send(msg []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}
func (c *ClientWebSocketAdapter) Receive() ([]byte, error) {
	_, msg, err := c.conn.ReadMessage()
	return msg, err
}
func (c *ClientWebSocketAdapter) HandleDisconnect() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Close()
}
