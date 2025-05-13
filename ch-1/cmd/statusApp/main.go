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
	"github.com/codyonesock/backend_learning/ch-1/internal/routes"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
	"github.com/codyonesock/backend_learning/ch-1/internal/storage"
	"github.com/codyonesock/backend_learning/ch-1/internal/users"
)

const (
	readTimeout         = 10 * time.Second
	writeTimeout        = 10 * time.Second
	idleTimeout         = 10 * time.Second
	sleepTimeout        = 5 * time.Second
	contextTimeout      = 15 * time.Minute
	authTokenExpiration = 24 * time.Hour
	saveInterval        = 1 * time.Minute
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
		zap.String("port", config.Port),
		zap.String("stream_url", config.StreamURL),
		zap.String("log_level", config.LogLevel),
		zap.Bool("use_scylla", config.UseScylla),
	)

	storageBackend := initializeStorage(config, logger)
	if scyllaStorage, ok := storageBackend.(*storage.ScyllaStorage); ok {
		defer scyllaStorage.Session.Close()
	}

	statsService := stats.NewStatsService(logger, storageBackend)
	statusService := status.NewStatusService(logger, statsService, sleepTimeout, contextTimeout)
	usersService := users.NewUserService(logger, config.JwtSecret, authTokenExpiration)

	setupStatsPersistence(statsService, logger, config.UseScylla, saveInterval)
	startServer(config, logger, statsService, statusService, usersService)
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

// initializeStorage sets up the storage backend (Scylla or in-memory).
//
//nolint:ireturn
func initializeStorage(config *config.Config, logger *zap.Logger) storage.Storage {
	if config.UseScylla {
		scyllaStorage, err := storage.NewScyllaStorage([]string{"localhost:9042"}, "stats_data", logger)
		if err != nil {
			logger.Fatal("Failed to initialize Scylla storage", zap.Error(err))
		}

		return scyllaStorage
	}

	logger.Info("Using in-memory storage")

	return storage.NewMemoryStorage()
}

// setupStatsPersistence will start saving stats data based on interval
// and if using scylla it will preload data.
func setupStatsPersistence(
	statsService *stats.Service,
	logger *zap.Logger,
	useScylla bool,
	saveInterval time.Duration,
) {
	statsService.StartPeriodicSave(saveInterval)

	if useScylla {
		if err := statsService.LoadStats(); err != nil {
			logger.Warn("Failed to load stats from Scylla", zap.Error(err))
		}
	}
}

// startServer starts the HTTP server with the provided services.
func startServer(
	config *config.Config,
	logger *zap.Logger,
	statsService *stats.Service,
	statusService *status.Service,
	usersService *users.Service,
) {
	r := chi.NewRouter()
	routes.RegisterRoutes(r, config.StreamURL, statsService, statusService, usersService)

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
