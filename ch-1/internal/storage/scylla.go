// Package storage - scyolla implementation.
package storage

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
)

// ScyllaStorage handles dependencies and config.
type ScyllaStorage struct {
	Session *gocql.Session
	Logger  *zap.Logger
}

// NewScyllaStorage returns a ScyllaStorage struct or error.
func NewScyllaStorage(hosts []string, keyspace string, logger *zap.Logger) (*ScyllaStorage, error) {
	const (
		maxAttempts = 10
		retryDelay  = 3 * time.Second
	)

	var (
		session *gocql.Session
		err     error
	)

	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		session, err = cluster.CreateSession()
		if err == nil {
			break
		}

		logger.Warn("Failed to connect to Scylla, retrying...",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		time.Sleep(retryDelay)
	}

	if err != nil {
		logger.Error("Failed to connect to Scylla after retries", zap.Error(err))
		return nil, fmt.Errorf("failed to create Scylla session after %d attempts: %w", maxAttempts, err)
	}

	return &ScyllaStorage{
		Session: session,
		Logger:  logger,
	}, nil
}

// SaveStats saves stats data.
func (s *ScyllaStorage) SaveStats(data *shared.Stats) error {
	query := `INSERT INTO stats (id, messages_consumed, distinct_users, bots_count, non_bots_count, distinct_server_urls)
                VALUES (?, ?, ?, ?, ?, ?)`

	id := gocql.TimeUUID()
	err := s.Session.Query(
		query,
		id,
		data.MessagesConsumed,
		data.DistinctUsers,
		data.BotsCount,
		data.NonBotsCount,
		data.DistinctServerURLs,
	).Exec()

	if err != nil {
		s.Logger.Error("Failed to save stats to Scylla", zap.Error(err))
		return fmt.Errorf("failed to execute query to save stats: %w", err)
	}

	s.Logger.Info("Stats saved to Scylla", zap.String("id", id.String()))

	return nil
}

// LoadStats returns stats.
func (s *ScyllaStorage) LoadStats() (*shared.Stats, error) {
	query := `SELECT
							messages_consumed,
							distinct_users,
							bots_count,
							non_bots_count,
							distinct_server_urls
						FROM stats LIMIT 1`

	stats := &shared.Stats{
		MessagesConsumed:   0,
		DistinctUsers:      make(map[string]int),
		BotsCount:          0,
		NonBotsCount:       0,
		DistinctServerURLs: make(map[string]int),
	}
	err := s.Session.Query(query).Scan(
		&stats.MessagesConsumed,
		&stats.DistinctUsers,
		&stats.BotsCount,
		&stats.NonBotsCount,
		&stats.DistinctServerURLs,
	)

	if stats.DistinctUsers == nil {
		stats.DistinctUsers = make(map[string]int)
	}
	if stats.DistinctServerURLs == nil {
		stats.DistinctServerURLs = make(map[string]int)
	}

	if err != nil {
		s.Logger.Error("Failed to load stats from Scylla", zap.Error(err))
		return nil, fmt.Errorf("failed to scan query result: %w", err)
	}

	s.Logger.Info("Stats loaded from Scylla")

	return stats, nil
}
