// Package main reads from wikimedia stream and produces messages to Redpanda.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/config"
	"github.com/codyonesock/backend_learning/ch-1/internal/logger"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

func main() {
	config := loadConfig()
	logger := initializeLogger(config)

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	logger.Info("Config loaded",
		zap.String("stream_url", config.StreamURL),
	)

	cl, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"), // TODO: Adjust for docker-compose later
		kgo.DefaultProduceTopic("wikimedia-changes"),
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
		logger.Info("Shutting down producer...")
		cancel()
	}()

	err = status.StreamAndProduce(ctx, config.StreamURL, cl, logger)
	if err != nil && ctx.Err() == nil {
		logger.Fatal("producer error", zap.Error(err))
	}

	logger.Info("Producer exited cleanly")
}

// loadConfig loads the config.
func loadConfig() *config.Config {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	return config
}

// initializeLogger sets up the zap logger.
func initializeLogger(config *config.Config) *zap.Logger {
	logger, err := logger.CreateLogger(config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}
