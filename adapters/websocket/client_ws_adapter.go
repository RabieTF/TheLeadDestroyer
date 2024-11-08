package websocket

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

	if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		HandleSendError("client", err)
		return err
	}
	return nil
}

func (c *ClientWebSocketAdapter) Receive() ([]byte, error) {
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		HandleReceiveError("client", err)
	}
	return msg, err
}

func (c *ClientWebSocketAdapter) HandleDisconnect() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if err := c.conn.Close(); err != nil {
		HandleDisconnectionError(err)
		return err
	}
	return nil
}
