// Package config is for wiring up config.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type config struct {
	Port      string `default:":7000"        envconfig:"PORT"`
	StreamURL string `envconfig:"STREAM_URL" required:"true"`
	LogLevel  string `default:"DEBUG"        envconfig:"LOG_LEVEL"`
}

// LoadConfig loads the application config.
func LoadConfig(logger *zap.Logger) (*config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Error("Error loading environment variables", zap.Error(err))
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	logger.Info("Config loaded",
		zap.String("port", cfg.Port),
		zap.String("stream_url", cfg.StreamURL),
		zap.String("log_level", cfg.LogLevel),
	)

	return &cfg, nil
}
