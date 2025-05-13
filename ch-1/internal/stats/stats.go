// Package stats can either return or update stats.
package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/storage"
)

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
	UpdateStats(rc shared.RecentChange)
	GetStats(w http.ResponseWriter) error
}

// Service handles dependencies.
type Service struct {
	Logger  *zap.Logger
	Mu      sync.Mutex
	Stats   *shared.Stats
	Storage storage.Storage
}

// NewStatsService create a new instance of Service.
func NewStatsService(l *zap.Logger, storage storage.Storage) *Service {
	return &Service{
		Logger: l,
		Mu:     sync.Mutex{},
		Stats: &shared.Stats{
			MessagesConsumed:   0,
			DistinctUsers:      map[string]int{},
			BotsCount:          0,
			NonBotsCount:       0,
			DistinctServerURLs: map[string]int{},
		},
		Storage: storage,
	}
}

// SaveStats saves the current stats.
func (s *Service) SaveStats() error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if err := s.Storage.SaveStats(s.Stats); err != nil {
		return fmt.Errorf("failed to save stats: %w", err)
	}

	return nil
}

// LoadStats loads the current stats.
func (s *Service) LoadStats() error {
	stats, err := s.Storage.LoadStats()
	if err != nil {
		return fmt.Errorf("failed to load stats: %w", err)
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.Stats = stats

	return nil
}

// StartPeriodicSave will peridically save stats data.
func (s *Service) StartPeriodicSave(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.SaveStats(); err != nil {
				s.Logger.Error("Failed to save stats periodically", zap.Error(err))
			}
		}
	}()
}

// Handler returns the router for /stats routes.
func (s *Service) Handler(statsService *Service) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		if err := statsService.GetStats(w); err != nil {
			s.Logger.Error("Error getting stats", zap.Error(err))
			http.Error(w, "Error getting stats", http.StatusInternalServerError)
		}
	})

	return r
}

// UpdateStats updates the Stats with the given RecentChange.
func (s *Service) UpdateStats(rc shared.RecentChange) {
	s.Mu.Lock()

	s.Stats.MessagesConsumed++
	s.Stats.DistinctUsers[rc.User]++
	s.Stats.DistinctServerURLs[rc.ServerURL]++

	if rc.Bot {
		s.Stats.BotsCount++
	} else {
		s.Stats.NonBotsCount++
	}

	s.Logger.Info("Stats updated", zap.String("user", rc.User), zap.Bool("bot", rc.Bot))
	s.Mu.Unlock()

	if err := s.SaveStats(); err != nil {
		s.Logger.Error("Failed to save stats after update", zap.Error(err))
	}
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
