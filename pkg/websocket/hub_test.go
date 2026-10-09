package websocket_test

import (
	"sync"
	"testing"

	wsPkg "joints-be/pkg/websocket"
)

type mockClient struct {
	mu       sync.Mutex
	received [][]byte
	closed   bool
}

func newMockClient() *mockClient {
	return &mockClient{received: make([][]byte, 0)}
}

func (c *mockClient) Send(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.received = append(c.received, message)
}

func (c *mockClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

func TestWebSocketHub_RegisterBroadcastUnregister(t *testing.T) {
	hub := wsPkg.NewHub(5)

	client1 := newMockClient()
	client2 := newMockClient()
	otherClient := newMockClient()

	// Register 2 clients to session-1
	if err := hub.Register("session-1", client1); err != nil {
		t.Fatalf("failed to register client1: %v", err)
	}
	if err := hub.Register("session-1", client2); err != nil {
		t.Fatalf("failed to register client2: %v", err)
	}

	// Register 1 client to session-2
	if err := hub.Register("session-2", otherClient); err != nil {
		t.Fatalf("failed to register otherClient: %v", err)
	}

	// Broadcast message to session-1
	payload := map[string]string{"event": "context_update", "concept": "Hukum Newton"}
	hub.Broadcast("session-1", payload)

	// Check client1 & client2 received message
	if len(client1.received) != 1 {
		t.Errorf("expected client1 to receive 1 message, got %d", len(client1.received))
	}
	if len(client2.received) != 1 {
		t.Errorf("expected client2 to receive 1 message, got %d", len(client2.received))
	}

	// Session isolation: otherClient in session-2 should NOT receive the message
	if len(otherClient.received) != 0 {
		t.Errorf("expected otherClient in session-2 to receive 0 messages, got %d", len(otherClient.received))
	}

	// Unregister client1
	hub.Unregister("session-1", client1)
	if !client1.closed {
		t.Errorf("expected client1 to be closed upon unregister")
	}

	// Broadcast again to session-1
	hub.Broadcast("session-1", payload)

	if len(client1.received) != 1 {
		t.Errorf("expected client1 to still have only 1 message after unregister, got %d", len(client1.received))
	}
	if len(client2.received) != 2 {
		t.Errorf("expected client2 to receive second message, got %d", len(client2.received))
	}

	// Close hub
	hub.Close()
	if !client2.closed || !otherClient.closed {
		t.Errorf("expected remaining clients to be closed upon hub.Close()")
	}
}

func TestWebSocketHub_MaxConnections(t *testing.T) {
	hub := wsPkg.NewHub(2)

	c1 := newMockClient()
	c2 := newMockClient()
	c3 := newMockClient()

	_ = hub.Register("s1", c1)
	_ = hub.Register("s1", c2)

	// 3rd client should be rejected due to limit
	err := hub.Register("s1", c3)
	if err != wsPkg.ErrMaxConnectionsReached {
		t.Errorf("expected ErrMaxConnectionsReached, got: %v", err)
	}
}

