package status_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/metrics"
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

type mockProducer struct {
	produced [][]byte
}

// Produce appends the produced record's value to the mockProducer's produced slice
// and invokes the callback to simulate successful production.
func (m *mockProducer) Produce(_ context.Context, record *kgo.Record, cb func(*kgo.Record, error)) {
	m.produced = append(m.produced, record.Value)
	cb(record, nil)
}

// TestStreamAndProduce sets up a mock HTTP server and producer, then tests that
// StreamAndProduce reads from the stream and produces at least one message.
func TestStreamAndProduce(t *testing.T) {
	t.Parallel()

	// simulate Wikimedia stream
	server := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(
			w,
			"data: {\"user\":\"blub\",\"bot\":false,\"server_url\":\"https://blub.com\"}\n"); err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}
	})
	ts := httptest.NewServer(server)

	defer ts.Close()

	mp := &mockProducer{
		produced: [][]byte{},
	}
	logger := zap.NewNop()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)

	defer cancel()

	m := metrics.NewProducerMetrics()

	err := status.StreamAndProduce(ctx, ts.URL, mp, logger, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mp.produced) == 0 {
		t.Errorf("expected at least one message produced")
	}
}
