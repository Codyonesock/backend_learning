package status_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

type MockStatsInterface struct {
	UpdatedChanges []shared.RecentChange
}

func (m *MockStatsInterface) UpdateStats(rc shared.RecentChange) {
	m.UpdatedChanges = append(m.UpdatedChanges, rc)
}

func (m *MockStatsInterface) GetStats(_ http.ResponseWriter) error {
	return nil // No-op - make the stats interface happy
}

// TestProcessStream sets up a mock stats interface and
// uses it to test processing that stream, verifying it at the end.
func TestProcessStream(t *testing.T) {
	t.Parallel()

	// We need a mock StatsInterface to pass to Status
	mockStats := &MockStatsInterface{
		UpdatedChanges: []shared.RecentChange{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(
			"data: {\"user\":\"blub_user\",\"bot\":false,\"server_url\":\"https://blub.com\"}\n"),
		); err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}
	}))
	defer server.Close()

	logger := zap.NewNop()
	service := status.NewStatusService(logger, mockStats, 1*time.Second, 5*time.Second)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	if err := service.ProcessStream(ctx, server.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockStats.UpdatedChanges) != 1 {
		t.Fatalf("expected 1 update, got %d", len(mockStats.UpdatedChanges))
	}

	rc := mockStats.UpdatedChanges[0]
	if rc.User != "blub_user" {
		t.Errorf("expected user to be 'blub_user', got '%s'", rc.User)
	}

	if rc.Bot {
		t.Errorf("expected bot to be false, got true")
	}

	if rc.ServerURL != "https://blub.com" {
		t.Errorf("expected server_url to be 'https://blub.com', got '%s'", rc.ServerURL)
	}
}
