package network

import (
	"context"
	"log/slog"
	"sync"

	"github.com/slimeyquest/server/internal/session"
)

// Hub manages active WebSocket connections.
type Hub struct {
	log      *slog.Logger
	sessions *session.Manager
	mu       sync.RWMutex
	conns    map[string]*Conn
	closed   bool
}

// NewHub creates a connection manager.
func NewHub(log *slog.Logger, sessions *session.Manager) *Hub {
	return &Hub{
		log:      log,
		sessions: sessions,
		conns:    make(map[string]*Conn),
	}
}

// Register adds a connection to the hub.
func (h *Hub) Register(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.conns[conn.ID()] = conn
	h.log.Info("websocket connected", "conn_id", conn.ID(), "active", len(h.conns))
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.conns[conn.ID()]; !ok {
		return
	}
	delete(h.conns, conn.ID())

	var token string
	var playerID int64
	if removed := h.sessions.UnbindConn(conn); removed != nil {
		token = removed.Token
		playerID = removed.PlayerID
	}

	h.log.Info("websocket disconnected",
		"conn_id", conn.ID(),
		"player_id", playerID,
		"token", token,
		"active", len(h.conns),
	)
}

// CloseAll closes every active connection during shutdown.
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for id, conn := range h.conns {
		conn.Close()
		delete(h.conns, id)
	}
	h.log.Info("websocket hub closed", "remaining", len(h.conns))
}

// Count returns the number of active connections.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

// Run is a placeholder for future hub-level background work.
func (h *Hub) Run(ctx context.Context) {
	<-ctx.Done()
}
