//go:build integration
// +build integration

package storage

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
)

func TestScyllaStorage_SaveAndLoadStats(t *testing.T) {
	t.Parallel()

	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("skipping integration test; set INTEGRATION=1 to run")
	}

	logger := zap.NewNop()
	hosts := []string{"localhost:9042"}
	keyspace := "stats_data"

	time.Sleep(5 * time.Second)

	storage, err := NewScyllaStorage(hosts, keyspace, logger)
	if err != nil {
		t.Fatalf("failed to connect to Scylla: %v", err)
	}
	defer storage.Session.Close()

	stats := &shared.Stats{
		MessagesConsumed:   42,
		DistinctUsers:      map[string]int{"blub": 1},
		BotsCount:          2,
		NonBotsCount:       40,
		DistinctServerURLs: map[string]int{"https://blub.com": 1},
	}

	if err := storage.Session.Query("TRUNCATE stats").Exec(); err != nil {
		t.Fatalf("failed to truncate stats table: %v", err)
	}

	if err := storage.SaveStats(stats); err != nil {
		t.Fatalf("failed to save stats: %v", err)
	}

	loaded, err := storage.LoadStats()
	if err != nil {
		t.Fatalf("failed to load stats: %v", err)
	}

	if loaded.MessagesConsumed != stats.MessagesConsumed {
		t.Errorf("expected %d, got %d", stats.MessagesConsumed, loaded.MessagesConsumed)
	}
}
