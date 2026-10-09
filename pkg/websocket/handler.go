package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"joints-be/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Izinkan CORS origin untuk WebSocket (dapat diperketat di konfigurasi prod)
		return true
	},
}

type Handler struct {
	hub Hub
}

func NewHandler(hub Hub) *Handler {
	return &Handler{hub: hub}
}

// HandleConnection menangani upgrade HTTP request ke WebSocket koneksi pada /ws/sessions/:sessionId.
func (h *Handler) HandleConnection(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		sessionID = c.Param("id")
	}
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, response.Error("sessionId parameter is required", nil))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewWSClient(h.hub, sessionID, conn)
	if err := h.hub.Register(sessionID, client); err != nil {
		_ = conn.WriteJSON(response.Error(err.Error(), nil))
		conn.Close()
		return
	}

	if wsCl, ok := client.(*wsClient); ok {
		wsCl.StartPumps()
	}
}

