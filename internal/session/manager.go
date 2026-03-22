package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager handles the lifecycle of Claude sessions.
// It is safe for concurrent use.
type Manager struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	storageDir string
}

// NewManager creates a new session manager that persists state to storageDir.
func NewManager(storageDir string) *Manager {
	return &Manager{
		sessions:   make(map[string]*Session),
		storageDir: storageDir,
	}
}

// Create creates a new session for the given key.
// It returns an error if a session with that key already exists.
// The returned *Session is a copy; mutating it does not affect the manager.
func (m *Manager) Create(key, workdir string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[key]; exists {
		return nil, fmt.Errorf("session already exists: %s", key)
	}

	now := time.Now()
	sess := &Session{
		Key:          key,
		Workdir:      workdir,
		State:        StateActive,
		CreatedAt:    now,
		LastActiveAt: now,
	}
	m.sessions[key] = sess
	cp := *sess
	return &cp, nil
}

// Get returns a copy of the session for the given key, or nil if not found.
// The returned *Session is a snapshot; mutating it does not affect the manager.
func (m *Manager) Get(key string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[key]
	if !ok {
		return nil
	}
	cp := *sess
	return &cp
}

// GetOrCreate returns a copy of the existing session for the key, or creates a
// new one with defaultWorkdir if none exists.
// The returned *Session is a snapshot; mutating it does not affect the manager.
func (m *Manager) GetOrCreate(key, defaultWorkdir string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess, exists := m.sessions[key]; exists {
		cp := *sess
		return &cp
	}

	now := time.Now()
	sess := &Session{
		Key:          key,
		Workdir:      defaultWorkdir,
		State:        StateActive,
		CreatedAt:    now,
		LastActiveAt: now,
	}
	m.sessions[key] = sess
	cp := *sess
	return &cp
}

// Clear resets the Claude session ID but preserves the workdir and keeps the
// session active. Use this when you want to start a fresh Claude conversation
// in the same working directory.
func (m *Manager) Clear(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[key]
	if !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	sess.ClaudeSession = ""
	sess.State = StateActive
	sess.LastActiveAt = time.Now()
	return nil
}

// Kill removes a session entirely.
func (m *Manager) Kill(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[key]; !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	delete(m.sessions, key)
	return nil
}

// SetWorkdir changes the working directory for an existing session.
func (m *Manager) SetWorkdir(key, workdir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[key]
	if !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	sess.Workdir = workdir
	sess.LastActiveAt = time.Now()
	return nil
}

// SetClaudeSession sets the Claude --resume session ID for an existing session.
func (m *Manager) SetClaudeSession(key, claudeSessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[key]
	if !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	sess.ClaudeSession = claudeSessionID
	sess.LastActiveAt = time.Now()
	return nil
}

// SetModel changes the model override for an existing session.
func (m *Manager) SetModel(key, model string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[key]
	if !exists {
		return fmt.Errorf("session not found: %s", key)
	}

	sess.Model = model
	return nil
}

// List returns copies of all sessions. The returned slice is a snapshot;
// mutating the returned sessions does not affect the manager's internal state.
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Session, 0, len(m.sessions))
	for _, sess := range m.sessions {
		cp := *sess
		list = append(list, &cp)
	}
	return list
}

// Save persists all sessions to storageDir/sessions.json.
// The write is atomic: data is written to a temporary file first, then renamed.
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	path := filepath.Join(m.storageDir, "sessions.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", tmpPath, path, err)
	}
	return nil
}

// Load reads sessions from storageDir/sessions.json. If the file does not
// exist, it is treated as a fresh start (no error, empty session map).
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := filepath.Join(m.storageDir, "sessions.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	sessions := make(map[string]*Session)
	if err := json.Unmarshal(data, &sessions); err != nil {
		return fmt.Errorf("failed to unmarshal sessions: %w", err)
	}

	m.sessions = sessions
	return nil
}

// Touch updates the LastActiveAt timestamp for the given session.
// If the session does not exist, it is a no-op.
func (m *Manager) Touch(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess, exists := m.sessions[key]; exists {
		sess.LastActiveAt = time.Now()
	}
}
