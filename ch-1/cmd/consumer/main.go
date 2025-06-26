package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/appInit"
	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/storage"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	config := appInit.MustLoadConfig()
	logger := appInit.MustInitLogger(config)

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	logger.Info("Config loaded",
		zap.String("stream_url", config.StreamURL),
	)

	storageBackend := appInit.MustInitStorage(config, logger)
	if scyllaStorage, ok := storageBackend.(*storage.ScyllaStorage); ok {
		defer scyllaStorage.Session.Close()
	}

	// statsService := stats.NewStatsService(logger, storageBackend)

	cl, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"), // TODO: Adjust for docker-compose
		kgo.ConsumerGroup("wikimedia-consumer-group"),
		kgo.ConsumeTopics("wikimedia-changes"),
	)

	if err != nil {
		logger.Fatal("failed to create Redpanda client", zap.Error(err))
	}

	defer cl.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down consumer...")
		cancel()
	}()

	logger.Info("Consumer started, waiting for messages...")

	for {
		if ctx.Err() != nil {
			break
		}

		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				logger.Warn("fetch error", zap.Error(err.Err))
			}
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			var rc shared.RecentChange
			if err := json.Unmarshal(record.Value, &rc); err != nil {
				logger.Warn("failed to unmarshal event", zap.Error(err))
				return
			}
			// err := statsService.UpdateStats(ctx, rc)
			// if err != nil {
			// 	logger.Warn("failed to update stats", zap.Error(err))
			// }
		})
	}

	logger.Info("Consumer exited cleanly")
}
