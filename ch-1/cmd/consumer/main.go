// Package main connets to Redpanda and consumes messages to process and update a stats db.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/appinit"
	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/storage"
)

func main() {
	config := appinit.MustLoadConfig()
	logger := appinit.MustInitLogger(config)

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()
	logger.Info("Config loaded", zap.String("stream_url", config.StreamURL))

	storageBackend := appinit.MustInitStorage(config, logger)
	if scyllaStorage, ok := storageBackend.(*storage.ScyllaStorage); ok {
		defer scyllaStorage.Session.Close()
	}

	statsService := stats.NewStatsService(logger, storageBackend)

	cl, err := setupKafkaClient()
	if err != nil {
		logger.Fatal("failed to create Redpanda client", zap.Error(err))
	}
	defer cl.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleShutdown(logger, cancel)

	logger.Info("Consumer started, waiting for messages...")

	processMessages(ctx, cl, logger, statsService)

	logger.Info("Consumer exited cleanly")
}

func setupKafkaClient() (*kgo.Client, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers("redpanda:9092"),
		kgo.ConsumerGroup("wikimedia-consumer-group"),
		kgo.ConsumeTopics("wikimedia-changes"),
		kgo.DisableAutoCommit(),
	)

	return cl, fmt.Errorf("error setting up redpanda client: %w", err)
}

func handleShutdown(logger *zap.Logger, cancel context.CancelFunc) {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down consumer...")
		cancel()
	}()
}

func processMessages(ctx context.Context, cl *kgo.Client, logger *zap.Logger, statsService *stats.Service) {
	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				logger.Warn("fetch error", zap.Error(err.Err))
			}

			continue
		}

		records := fetches.Records()
		if len(records) == 0 {
			continue
		}

		var batch []shared.RecentChange

		for _, record := range records {
			var rc shared.RecentChange
			if err := json.Unmarshal(record.Value, &rc); err != nil {
				logger.Warn("failed to unmarshal event", zap.Error(err))
				return
			}

			batch = append(batch, rc)
		}

		for _, rc := range batch {
			statsService.UpdateStats(rc)
		}

		if err := cl.CommitRecords(ctx, records...); err != nil {
			logger.Warn("failed to commit offsets", zap.Error(err))
		}
	}
}
