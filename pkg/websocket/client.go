package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type wsClient struct {
	hub       Hub
	sessionID string
	conn      *websocket.Conn
	send      chan []byte
	once      sync.Once
	closed    chan struct{}
}

func NewWSClient(hub Hub, sessionID string, conn *websocket.Conn) Client {
	return &wsClient{
		hub:       hub,
		sessionID: sessionID,
		conn:      conn,
		send:      make(chan []byte, 64),
		closed:    make(chan struct{}),
	}
}

func (c *wsClient) Send(message []byte) {
	select {
	case <-c.closed:
		return
	case c.send <- message:
	default:
		log.Printf("[websocket] client buffer full session_id=%s, dropping message", c.sessionID)
	}
}

func (c *wsClient) Close() {
	c.once.Do(func() {
		close(c.closed)
		_ = c.conn.Close()
	})
}

// StartPumps memulai pump goroutine untuk membaca (handling disconnect/ping) dan menulis pesan.
func (c *wsClient) StartPumps() {
	go c.writePump()
	go c.readPump()
}

func (c *wsClient) readPump() {
	defer func() {
		c.hub.Unregister(c.sessionID, c)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[websocket] unexpected client close session_id=%s error=%v", c.sessionID, err)
			}
			break
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.hub.Unregister(c.sessionID, c)
	}()

	for {
		select {
		case <-c.closed:
			return
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Drain queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

