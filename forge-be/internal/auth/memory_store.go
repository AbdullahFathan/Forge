package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.Mutex
	refresh map[string]refreshRecord
	family  map[string]struct{}
	login   map[string]int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		refresh: map[string]refreshRecord{},
		family:  map[string]struct{}{},
		login:   map[string]int{},
	}
}

func (s *MemoryStore) SaveRefresh(_ context.Context, token, userID, familyID string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refresh[token] = refreshRecord{UserID: userID, FamilyID: familyID}
	s.family[familyID] = struct{}{}
	return nil
}

func (s *MemoryStore) GetRefresh(_ context.Context, token string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.refresh[token]
	if !ok {
		return "", "", fmt.Errorf("not found")
	}
	return rec.UserID, rec.FamilyID, nil
}

func (s *MemoryStore) DeleteRefresh(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refresh, token)
	return nil
}

func (s *MemoryStore) InvalidateFamily(_ context.Context, familyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.family, familyID)
	return nil
}

func (s *MemoryStore) FamilyValid(_ context.Context, familyID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.family[familyID]
	return ok, nil
}

func (s *MemoryStore) AllowLogin(_ context.Context, ip string, limit int, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.login[ip]++
	return s.login[ip] <= limit, nil
}
