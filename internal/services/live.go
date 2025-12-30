package services

import (
	"4connect/internal/models"
	"errors"
	"sync"
)

type SessionManager struct {
	sessions map[string]*models.Session
	mu       sync.Mutex
}

func MakeSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[string]*models.Session), mu: sync.Mutex{}}
}

func (manager *SessionManager) CreateSession(matchID string) *models.Session {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	session := models.MakeSession()
	manager.sessions[matchID] = session
	return session
}

func (manager *SessionManager) GetSession(matchID string) (*models.Session, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	session, ok := manager.sessions[matchID]
	if !ok {
		return nil, errors.New("Session does not exist")
	}
	return session, nil
}
