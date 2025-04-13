// Package stats can either return or update stats.
package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/models"
)

// Stats holds the core data that comes from Wikimedia.
type Stats struct {
	MessagesConsumed   int            `json:"messages_consumed"`
	DistinctUsers      map[string]int `json:"-"`
	BotsCount          int            `json:"bots_count"`
	NonBotsCount       int            `json:"non_bots_count"`
	DistinctServerURLs map[string]int `json:"-"`
}

// Response returns all the counts in ints.
type Response struct {
	MessagesConsumed       int `json:"messages_consumed"`
	DistinctUsersCount     int `json:"distinct_users"`
	BotsCount              int `json:"bots_count"`
	NonBotsCount           int `json:"non_bots_count"`
	DistinctServerURLCount int `json:"distinct_server_urls"`
}

// ServiceInterface will be used to update the stats.
type ServiceInterface interface {
	UpdateStats(rc models.RecentChange)
	GetStats(w http.ResponseWriter) error
}

// Service handles dependencies.
type Service struct {
	Logger *zap.Logger
	Mu     sync.Mutex
	Stats  Stats
}

// NewStatsService create a new instance of Service.
func NewStatsService(l *zap.Logger) *Service {
	return &Service{
		Logger: l,
		Mu:     sync.Mutex{},
		Stats: Stats{
			MessagesConsumed:   0,
			DistinctUsers:      map[string]int{},
			BotsCount:          0,
			NonBotsCount:       0,
			DistinctServerURLs: map[string]int{},
		},
	}
}

// UpdateStats updates the Stats with the given RecentChange.
func (s *Service) UpdateStats(rc models.RecentChange) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.Stats.MessagesConsumed++
	s.Stats.DistinctUsers[rc.User]++
	s.Stats.DistinctServerURLs[rc.ServerURL]++

	if rc.Bot {
		s.Stats.BotsCount++
	} else {
		s.Stats.NonBotsCount++
	}

	s.Logger.Info("Stats updated", zap.String("user", rc.User), zap.Bool("bot", rc.Bot))
}

// GetStats returns the current StatsResponse.
func (s *Service) GetStats(w http.ResponseWriter) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	response := Response{
		MessagesConsumed:       s.Stats.MessagesConsumed,
		DistinctUsersCount:     len(s.Stats.DistinctUsers),
		BotsCount:              s.Stats.BotsCount,
		NonBotsCount:           s.Stats.NonBotsCount,
		DistinctServerURLCount: len(s.Stats.DistinctServerURLs),
	}

	w.Header().Set("Content-Type", "application/json")

	statsJSON, err := json.Marshal(response)

	if err != nil {
		s.Logger.Error("Failed to marshal response", zap.Error(err))
		return fmt.Errorf("failed to marshal response %w", err)
	}

	if _, err := w.Write(statsJSON); err != nil {
		s.Logger.Error("Failed to write stats response", zap.Error(err))
	}

	s.Logger.Info("Current stats", zap.Any("response", response))

	return nil
}
