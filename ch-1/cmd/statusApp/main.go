// Package main initializes the logger and routing.
// It also loads a config to configure the port and stream url
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

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

type config struct {
	Port      string `default:":7000"        envconfig:"PORT"`
	StreamURL string `envconfig:"STREAM_URL" required:"true"`
}

func loadConfig(logger *zap.Logger) (*config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("Error loading environment variables", zap.Error(err))
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	logger.Info("Config loaded",
		zap.String("port", cfg.Port),
		zap.String("stream_url", cfg.StreamURL),
	)

	return &cfg, nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	config, err := loadConfig(logger)
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
		return
	}

	r := chi.NewRouter()

	var (
		statsService  stats.ServiceInterface  = stats.NewStatsService(logger)
		statusService status.ServiceInterface = status.NewStatusService(logger, statsService, sleepTimeout, contextTimeout)
	)

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
