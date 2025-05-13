// Package storage - in-memory implementation.
package storage

import (
	"sync"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
type MemoryStorage struct {
	mu    sync.Mutex
	stats *shared.Stats
}

// NewMemoryStorage creates a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mu: sync.Mutex{},
		stats: &shared.Stats{
			MessagesConsumed:   0,
			DistinctUsers:      map[string]int{},
			BotsCount:          0,
			NonBotsCount:       0,
			DistinctServerURLs: map[string]int{},
		},
	}
}

// SaveStats saves stats data in memory.
func (m *MemoryStorage) SaveStats(stat *shared.Stats) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = stat

	return nil
}

// LoadStats loads stats data from memory.
func (m *MemoryStorage) LoadStats() (*shared.Stats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stats, nil
}
