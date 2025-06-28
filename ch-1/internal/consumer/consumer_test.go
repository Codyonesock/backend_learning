// Package consumer_test provides tests for the consumer package.
package consumer_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/codyonesock/backend_learning/ch-1/internal/consumer"
	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ctxKey string

// mockKafkaClient is a simple mock KafkaClient.
type mockKafkaClient struct {
	records   []*kgo.Record
	committed bool
}

func (f *mockKafkaClient) PollFetches(ctx context.Context) kgo.Fetches {
	// Cancel the context so the consumer loop exits
	if cancelFunc := ctx.Value(ctxKey("cancelFunc")); cancelFunc != nil {
		if cf, ok := cancelFunc.(func()); ok {
			cf()
		}
	}

	fetches := kgo.Fetches{}

	return fetches
}

func (f *mockKafkaClient) CommitRecords(_ context.Context, _ ...*kgo.Record) error {
	f.committed = true
	return nil
}

// mockStatsUpdater is a mock of the StatsUpdater interface.
type mockStatsUpdater struct {
	calls []shared.RecentChange
}

func (f *mockStatsUpdater) UpdateStats(rc shared.RecentChange) {
	f.calls = append(f.calls, rc)
}

// TestProcessMessages tests that a message is processed,
// stats are updated, and the offset is committed.
func TestProcessMessages(t *testing.T) {
	t.Parallel()

	rc := shared.RecentChange{User: "blubuser", ServerURL: "", Bot: false}

	val, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("failed to marshal rc: %v", err)
	}

	rec := &kgo.Record{Value: val}

	ctx, cancel := context.WithCancel(t.Context())
	ctx = context.WithValue(ctx, ctxKey("cancelFunc"), cancel)

	client := &mockKafkaClient{records: []*kgo.Record{rec}, committed: false}
	stats := &mockStatsUpdater{calls: []shared.RecentChange{}}
	logger := zaptest.NewLogger(t)

	go func() {
		consumer.ProcessMessages(ctx, client, logger, stats)
	}()

	// trigger cancel after a moment
	time.Sleep(100 * time.Millisecond)
	cancel()

	// wait for consumer to cancel
	time.Sleep(50 * time.Millisecond)

	if !client.committed {
		t.Errorf("expected CommitRecords to be called")
	}

	if len(stats.calls) != 1 {
		t.Fatalf("expected 1 call to UpdateStats, got %d", len(stats.calls))
	}

	if stats.calls[0].User != "blubuser" {
		t.Errorf("expected user 'blubuser', got '%s'", stats.calls[0].User)
	}
}
