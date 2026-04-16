package api

import (
	"crypto/rand"
	"fmt"
	"sync"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionDTO
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*SessionDTO)}
}

func (s *SessionStore) Create(data *SessionDTO) string {
	id := newUUID()
	data.ID = id
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = data
	return id
}

func (s *SessionStore) Get(id string) (*SessionDTO, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

func (s *SessionStore) Update(id string, data *SessionDTO) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return false
	}
	s.sessions[id] = data
	return true
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
