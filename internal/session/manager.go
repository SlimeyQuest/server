// Package session manages in-memory player connection sessions.
//
// Replacement policy: newest connection wins. Each player_id has at most one
// active session. Bind removes any prior session for a different connection and
// the login layer closes the replaced connection immediately.
package session

import (
	"sync"

	"github.com/google/uuid"
)

// LiveConn is the connection handle stored in a session.
type LiveConn interface {
	ID() string
	Close()
}

// Session binds an authenticated player to a live connection.
type Session struct {
	Token    string
	PlayerID int64
	Conn     LiveConn
}

// Manager tracks in-memory player sessions.
type Manager struct {
	mu         sync.Mutex
	byToken    map[string]*Session
	byPlayerID map[int64]*Session
	byConnID   map[string]*Session
}

// NewManager creates an in-memory session manager.
func NewManager() *Manager {
	return &Manager{
		byToken:    make(map[string]*Session),
		byPlayerID: make(map[int64]*Session),
		byConnID:   make(map[string]*Session),
	}
}

// Bind attaches a player to a connection, replacing any prior active session.
// Returns the new session and the replaced session when applicable.
func (m *Manager) Bind(playerID int64, conn LiveConn) (*Session, *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var replaced *Session
	if existing, ok := m.byPlayerID[playerID]; ok && existing.Conn.ID() != conn.ID() {
		replaced = existing
		m.removeLocked(existing)
	}

	session := &Session{
		Token:    uuid.NewString(),
		PlayerID: playerID,
		Conn:     conn,
	}

	m.byToken[session.Token] = session
	m.byPlayerID[playerID] = session
	m.byConnID[conn.ID()] = session
	return session, replaced
}

// UnbindConn removes the session for a connection when it is still active.
func (m *Manager) UnbindConn(conn LiveConn) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byConnID[conn.ID()]
	if !ok {
		return nil
	}

	current, ok := m.byPlayerID[session.PlayerID]
	if !ok || current.Conn.ID() != conn.ID() {
		delete(m.byConnID, conn.ID())
		return session
	}

	m.removeLocked(session)
	return session
}

// GetByPlayerID returns the active session for a player.
func (m *Manager) GetByPlayerID(playerID int64) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byPlayerID[playerID]
	return session, ok
}

// Validate reports whether the token is the active session for the player.
func (m *Manager) Validate(playerID int64, token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byPlayerID[playerID]
	if !ok {
		return false
	}
	return session.Token == token
}

// GetByToken returns a session by token.
func (m *Manager) GetByToken(token string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byToken[token]
	return session, ok
}

func (m *Manager) removeLocked(session *Session) {
	delete(m.byToken, session.Token)
	if current, ok := m.byPlayerID[session.PlayerID]; ok && current.Token == session.Token {
		delete(m.byPlayerID, session.PlayerID)
	}
	delete(m.byConnID, session.Conn.ID())
}
