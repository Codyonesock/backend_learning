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

	"github.com/codyonesock/backend_learning/ch-1/internal/appinit"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

func main() {
	config := appinit.MustLoadConfig()
	logger := appinit.MustInitLogger(config)

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	logger.Info("Config loaded",
		zap.String("stream_url", config.StreamURL),
	)

	cl, err := kgo.NewClient(
		kgo.SeedBrokers("redpanda:9092"),
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
