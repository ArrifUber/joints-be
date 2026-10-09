package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
)

var ErrMaxConnectionsReached = errors.New("maximum websocket connections reached")

// Client mendefinisikan antarmuka untuk koneksi websocket individual.
type Client interface {
	Send(message []byte)
	Close()
}

// Hub mendefinisikan antarmuka pengelola koneksi WebSocket realtime per session.
type Hub interface {
	Register(sessionID string, client Client) error
	Unregister(sessionID string, client Client)
	Broadcast(sessionID string, message any)
	Close()
}

type hubImpl struct {
	mu             sync.RWMutex
	sessions       map[string]map[Client]bool
	totalClients   int
	maxConnections int
}

func NewHub(maxConnections int) Hub {
	if maxConnections <= 0 {
		maxConnections = 100
	}
	return &hubImpl{
		sessions:       make(map[string]map[Client]bool),
		maxConnections: maxConnections,
	}
}

func (h *hubImpl) Register(sessionID string, client Client) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.totalClients >= h.maxConnections {
		log.Printf("[websocket] max connections (%d) reached, rejecting client", h.maxConnections)
		return ErrMaxConnectionsReached
	}

	if _, exists := h.sessions[sessionID]; !exists {
		h.sessions[sessionID] = make(map[Client]bool)
	}

	h.sessions[sessionID][client] = true
	h.totalClients++
	log.Printf("[websocket] client registered session_id=%s total_clients=%d", sessionID, h.totalClients)
	return nil
}

func (h *hubImpl) Unregister(sessionID string, client Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.sessions[sessionID]; exists {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			h.totalClients--
			client.Close()
			log.Printf("[websocket] client unregistered session_id=%s total_clients=%d", sessionID, h.totalClients)
		}
		if len(clients) == 0 {
			delete(h.sessions, sessionID)
		}
	}
}

func (h *hubImpl) Broadcast(sessionID string, message any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.sessions[sessionID]
	if !exists || len(clients) == 0 {
		log.Printf("[websocket] no active clients for session_id=%s, skipping broadcast", sessionID)
		return
	}

	raw, err := json.Marshal(message)
	if err != nil {
		log.Printf("[websocket] failed to serialize broadcast message: %v", err)
		return
	}

	for client := range clients {
		client.Send(raw)
	}
	log.Printf("[websocket] broadcasted event to %d client(s) on session_id=%s", len(clients), sessionID)
}

func (h *hubImpl) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("[websocket] closing all websocket connections (total=%d)...", h.totalClients)
	for sessionID, clients := range h.sessions {
		for client := range clients {
			client.Close()
		}
		delete(h.sessions, sessionID)
	}
	h.totalClients = 0
}

