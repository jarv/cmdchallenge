package store

import (
	"fmt"
	"sync"
)

type MemStore struct {
	mu      sync.Mutex
	results map[string]*CmdStore
}

func NewMemStore() (CmdStorer, error) {
	m := &MemStore{}
	m.results = make(map[string]*CmdStore)
	return m, nil
}

func (m *MemStore) TopCmdsForSlug(slug string) ([]string, error) {
	return make([]string, 0), nil
}

func (m *MemStore) GetResult(cmd, slug string, version int) (*CmdStore, error) {
	if !m.hasResult(genKey(&cmd, &slug, &version)) {
		return nil, ErrResultNotFound
	}

	return m.results[genKey(&cmd, &slug, &version)], nil
}

func (m *MemStore) CreateResult(s *CmdStore) error {
	if m.hasResult(genKey(s.Cmd, s.Slug, s.Version)) {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.results[genKey(s.Cmd, s.Slug, s.Version)] = s
	return nil
}

func (m *MemStore) IncrementResult(cmd, slug string, version int) error {
	return nil
}

func (m *MemStore) hasResult(key string) bool {
	if _, found := m.results[key]; found {
		return true
	} else {
		return false
	}
}

func genKey(cmd, slug *string, version *int) string {
	return fmt.Sprintf("%s-%s-%d", *cmd, *slug, *version)
}
