// Package consumer provides logic to consume messages from Redpanda and update application statistics.
package consumer

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/codyonesock/backend_learning/ch-1/internal/metrics"
	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	wikimedia "github.com/codyonesock/backend_learning/ch-6/proto"
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

const updateIntTime = 5 * time.Second

// ProcessMessages consumes messages from Redpanda, updates statistics, and commits offsets.
// It processes messages in batches and handles errors and acknowledgements.
//
// Resilience note:
// Commit the offsets after stats are updated and saved via UpdateStats batching.
// If the app crashes or is restarted before the commit, Kafka should process the messages.
func ProcessMessages(
	ctx context.Context,
	cl KafkaClient,
	logger *zap.Logger,
	statsService StatsUpdater,
	metrics *metrics.ConsumerMetrics,
) {
	// Note: I only added this because the volume spam is annoying.
	updateInterval := updateIntTime
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
			metrics.EventsProcessedFailed.Add(float64(len(errs)))

			continue
		}

		records := fetches.Records()
		if len(records) == 0 {
			continue
		}

		metrics.EventsConsumed.Add(float64(len(records)))

		batch := unmarshalRecords(records, logger)

		now := time.Now()
		if now.After(nextUpdate) {
			for _, rc := range batch {
				statsService.UpdateStats(rc)
				metrics.EventsProcessedSuccess.Inc()
			}

			nextUpdate = now.Add(updateInterval)
		}

		if err := cl.CommitRecords(ctx, records...); err != nil {
			logger.Warn("failed to commit offsets", zap.Error(err))
			metrics.EventsProcessedFailed.Inc()
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
		var pb wikimedia.RecentChange
		if err := proto.Unmarshal(record.Value, &pb); err != nil {
			logger.Warn("failed to unmarshal protobuf event", zap.Error(err))
			continue
		}

		rc := shared.RecentChange{
			User:      pb.GetUser(),
			Bot:       pb.GetBot(),
			ServerURL: pb.GetServerUrl(),
		}
		batch = append(batch, rc)
	}

	return batch
}
