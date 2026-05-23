// Package session manages in-memory player sessions.
//
// Replacement policy: newest session wins. Each player_id has at most one
// active session token.
package session

import (
	"sync"

	"github.com/google/uuid"
)

// Session binds an authenticated player to a session token.
type Session struct {
	Token    string
	PlayerID int64
}

// Manager tracks in-memory player sessions.
type Manager struct {
	mu         sync.Mutex
	byToken    map[string]*Session
	byPlayerID map[int64]*Session
}

// NewManager creates an in-memory session manager.
func NewManager() *Manager {
	return &Manager{
		byToken:    make(map[string]*Session),
		byPlayerID: make(map[int64]*Session),
	}
}

// Bind attaches a player to a new session, replacing any prior active session.
// Returns the new session and the replaced session when applicable.
func (m *Manager) Bind(playerID int64) (*Session, *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var replaced *Session
	if existing, ok := m.byPlayerID[playerID]; ok {
		replaced = existing
		m.removeLocked(existing)
	}

	session := &Session{
		Token:    uuid.NewString(),
		PlayerID: playerID,
	}

	m.byToken[session.Token] = session
	m.byPlayerID[playerID] = session
	return session, replaced
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
}
