// Package main initializes the logger and routing.
// It also loads a config to configure the port and stream url
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/config"
	"github.com/codyonesock/backend_learning/ch-1/internal/logger"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	idleTimeout    = 10 * time.Second
	sleepTimeout   = 5 * time.Second
	contextTimeout = 10 * time.Second
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}

	logger, err := logger.CreateLogger(config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	logger.Info("Config loaded",
		zap.String("port", config.Port),
		zap.String("stream_url", config.StreamURL),
		zap.String("log_level", config.LogLevel),
	)

	statsService := stats.NewStatsService(logger)
	statusService := status.NewStatusService(logger, statsService, sleepTimeout, contextTimeout)

	r := chi.NewRouter()

	r.Get("/status", func(w http.ResponseWriter, _ *http.Request) {
		if err := statusService.ProcessStream(config.StreamURL); err != nil {
			http.Error(w, "Error processing stream", http.StatusInternalServerError)
		}
	})

	r.Get("/stats", func(w http.ResponseWriter, _ *http.Request) {
		if err := statsService.GetStats(w); err != nil {
			http.Error(w, "Error getting stats", http.StatusInternalServerError)
		}
	})

	logger.Info("Server running", zap.String("port", config.Port))
	server := &http.Server{
		Addr:         config.Port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
}
