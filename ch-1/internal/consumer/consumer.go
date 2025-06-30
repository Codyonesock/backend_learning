// Package consumer provides logic to consume messages from Redpanda and update application statistics.
package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
)

// KafkaClient abstracts the Redpanda client for polling and committing records.
type KafkaClient interface {
	PollFetches(ctx context.Context) kgo.Fetches
	CommitRecords(ctx context.Context, records ...*kgo.Record) error
}

// StatsUpdater defines an interface for updating statistics from consumed messages.
type StatsUpdater interface {
	UpdateStats(rc shared.RecentChange)
}

// ProcessMessages consumes messages from Redpanda, updates statistics, and commits offsets.
// It processes messages in batches and handles errors and acknowledgements.
func ProcessMessages(ctx context.Context, cl KafkaClient, logger *zap.Logger, statsService StatsUpdater) {
	// Note: I only added this because the volume spam is annoying.
	updateInterval := 5 * time.Second
	nextUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			logFetchErrors(logger, errs)
			continue
		}

		records := fetches.Records()
		if len(records) == 0 {
			continue
		}

		batch := unmarshalRecords(records, logger)

		now := time.Now()
		if now.After(nextUpdate) {
			for _, rc := range batch {
				statsService.UpdateStats(rc)
			}
			nextUpdate = now.Add(updateInterval)
		}

		if err := cl.CommitRecords(ctx, records...); err != nil {
			logger.Warn("failed to commit offsets", zap.Error(err))
		}
	}
}

func logFetchErrors(logger *zap.Logger, errs []kgo.FetchError) {
	for _, err := range errs {
		logger.Warn("fetch error", zap.Error(err.Err))
	}
}

func unmarshalRecords(records []*kgo.Record, logger *zap.Logger) []shared.RecentChange {
	batch := make([]shared.RecentChange, 0, len(records))

	for _, record := range records {
		var rc shared.RecentChange
		if err := json.Unmarshal(record.Value, &rc); err != nil {
			logger.Warn("failed to unmarshal event", zap.Error(err))
			continue
		}

		batch = append(batch, rc)
	}

	return batch
}
