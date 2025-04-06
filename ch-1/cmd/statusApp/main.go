package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

type Config struct {
	Port      string `json:"port"`
	StreamURL string `json:"stream_url"`
}

func LoadConfig(filename string, logger *zap.Logger) (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		logger.Error("Error reading config file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
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
		fmt.Fprintf(os.Stderr, "Failed to initalize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting StatusApp")

	config, err := LoadConfig("config.json", logger)
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
		return
	}

	r := chi.NewRouter()

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		status.GetStatus(w, r, config.StreamURL, logger)
	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats.GetStats(w, r, logger)
	})

	logger.Info("Server running", zap.String("port", config.Port))
	if err := http.ListenAndServe(config.Port, r); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
		os.Exit(1)
	}
}

//! 1. Update to service pattern (receiver methods)
//! 2. Fix linting issues :')
//! 3. Fix the tests :D
