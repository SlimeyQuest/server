// Package session manages in-memory player sessions.
//
// Replacement policy: newest session wins. Each player_id has at most one
// active session. Interface adapters may use the replaced handle to close or
// notify an old transport connection.
package session

import (
	"sync"

	"github.com/google/uuid"
)

// Binding identifies an interface-layer session owner without exposing transport details.
type Binding struct {
	ID     string
	Handle any
}

// Session binds an authenticated player to a session token and interface binding.
type Session struct {
	Token    string
	PlayerID int64
	Binding  Binding
}

// Manager tracks in-memory player sessions.
type Manager struct {
	mu         sync.Mutex
	byToken    map[string]*Session
	byPlayerID map[int64]*Session
	byBinding  map[string]*Session
}

// NewManager creates an in-memory session manager.
func NewManager() *Manager {
	return &Manager{
		byToken:    make(map[string]*Session),
		byPlayerID: make(map[int64]*Session),
		byBinding:  make(map[string]*Session),
	}
}

// Bind attaches a player to an interface binding, replacing any prior active session.
// Returns the new session and the replaced session when applicable.
func (m *Manager) Bind(playerID int64, binding Binding) (*Session, *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var replaced *Session
	if existing, ok := m.byPlayerID[playerID]; ok && existing.Binding.ID != binding.ID {
		replaced = existing
		m.removeLocked(existing)
	}

	session := &Session{
		Token:    uuid.NewString(),
		PlayerID: playerID,
		Binding:  binding,
	}

	m.byToken[session.Token] = session
	m.byPlayerID[playerID] = session
	m.byBinding[binding.ID] = session
	return session, replaced
}

// Unbind removes the session for a binding when it is still active.
func (m *Manager) Unbind(bindingID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byBinding[bindingID]
	if !ok {
		return nil
	}

	current, ok := m.byPlayerID[session.PlayerID]
	if !ok || current.Binding.ID != bindingID {
		delete(m.byBinding, bindingID)
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
	delete(m.byBinding, session.Binding.ID)
}
