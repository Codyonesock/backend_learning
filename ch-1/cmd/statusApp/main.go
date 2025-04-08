// Package main initializes the logger and routing.
// It also loads a config to configure the port and stream url
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

type config struct {
	Port      string `json:"port"`
	StreamURL string `json:"stream_url"`
}

func loadConfig(filename string, logger *zap.Logger) (*config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		logger.Error("Error reading config file", zap.String("filename", filename), zap.Error(err))

		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config config
	err = json.Unmarshal(data, &config)

	if err != nil {
		logger.Error("Error unmarshalling JSON", zap.Error(err))

		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	logger.Info("Config loaded", zap.String("port", config.Port), zap.String("stream_url", config.StreamURL))

	return &config, nil
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

	config, err := loadConfig("config.json", logger)

	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))

		return
	}

	r := chi.NewRouter()

	var statusService status.StatusServiceInterface
	statusService = status.NewStatusService(logger)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		err := statusService.ProcessStream(config.StreamURL)
		if err != nil {
			http.Error(w, "Error processing stream", http.StatusInternalServerError)
		}
	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats.GetStats(w, r, logger)
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

//! 1. Update to service pattern (receiver methods)
//! 2. Fix linting issues :')
//! 3. Fix the tests (Tests should stil pass with -race flag) :D
