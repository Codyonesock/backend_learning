// Package main connets to Redpanda and consumes messages to process and update a stats db.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/appinit"
	"github.com/codyonesock/backend_learning/ch-1/internal/consumer"
	"github.com/codyonesock/backend_learning/ch-1/internal/metrics"
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

	cm := metrics.NewConsumerMetrics()
	metrics.StartServer(":2112")

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

	consumer.ProcessMessages(ctx, cl, logger, statsService, cm)

	logger.Info("Consumer exited cleanly")
}

func setupKafkaClient() (*kgo.Client, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers("redpanda:9092"),
		kgo.ConsumerGroup("wikimedia-consumer-group"),
		kgo.ConsumeTopics("wikimedia-changes-proto"),
		kgo.DisableAutoCommit(),
	)

	if err != nil {
		return cl, fmt.Errorf("error setting up redpanda client: %w", err)
	}

	return cl, nil
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
