// package appInit
package appInit

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/config"
	"github.com/codyonesock/backend_learning/ch-1/internal/logger"
	"github.com/codyonesock/backend_learning/ch-1/internal/storage"
)

// MustLoadConfig loads config or exits.
func MustLoadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// MustInitLogger initializes logger or exits.
func MustInitLogger(cfg *config.Config) *zap.Logger {
	log, err := logger.CreateLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	return log
}

// MustInitStorage initializes storage backend or uses in memory.
func MustInitStorage(cfg *config.Config, log *zap.Logger) storage.Storage {
	if cfg.UseScylla {
		scyllaStorage, err := storage.NewScyllaStorage([]string{"scylla:9042"}, "stats_data", log)
		if err != nil {
			log.Fatal("Failed to initialize Scylla storage", zap.Error(err))
		}
		return scyllaStorage
	}
	log.Info("Using in-memory storage")
	return storage.NewMemoryStorage()
}
