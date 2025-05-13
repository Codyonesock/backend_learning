// Package storage handles read and write.
package storage

import "github.com/codyonesock/backend_learning/ch-1/internal/shared"

// Storage defines the interface for storage backends.
type Storage interface {
	SaveStats(stat *shared.Stats) error
	LoadStats() (*shared.Stats, error)
}
