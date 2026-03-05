package state

import (
	"fmt"
	"sync"
)

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string][]byte)}
}

func (s *MemoryStore) Save(key string, data []byte) error {
	if key == "" {
		return fmt.Errorf("state key is required")
	}

	copyBytes := make([]byte, len(data))
	copy(copyBytes, data)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = copyBytes
	return nil
}

func (s *MemoryStore) Load(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("state key is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	raw, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}

	copyBytes := make([]byte, len(raw))
	copy(copyBytes, raw)
	return copyBytes, nil
}
