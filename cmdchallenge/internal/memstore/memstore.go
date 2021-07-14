package memstore

import (
	"sync"

	"gitlab.com/jarv/cmdchallenge/internal/runner"
)

type MemStore struct {
	sync.Mutex
	results map[string]*runner.RunnerResult
}

func New() (runner.RunnerResultStorer, error) {
	m := &MemStore{}
	m.results = make(map[string]*runner.RunnerResult)
	return m, nil
}

func (m *MemStore) TopCmdsForSlug(slug string) ([]string, error) {
	return make([]string, 0), nil
}

func (m *MemStore) GetResult(fingerprint string) (*runner.RunnerResult, error) {
	if !m.hasResult(fingerprint) {
		return nil, runner.ErrResultNotFound
	}

	return m.results[fingerprint], nil
}

func (m *MemStore) CreateResult(fingerprint, cmd, slug string, version int, result *runner.RunnerResult) error {
	if m.hasResult(fingerprint) {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	m.results[fingerprint] = result
	return nil
}

func (m *MemStore) hasResult(fingerprint string) bool {
	if _, found := m.results[fingerprint]; found {
		return true
	} else {
		return false
	}
}
