// Package config is for wiring up config.
package config

import (
	"errors"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

var errInvalidStreamURL = errors.New("STREAM_URL is required but not set")

// Config is your config.
type Config struct {
	Port      string `default:":7000"        envconfig:"PORT"`
	StreamURL string `envconfig:"STREAM_URL" required:"true"`
	LogLevel  string `default:"INFO"         envconfig:"LOG_LEVEL"`
	JwtSecret string `envconfig:"JWT_SECRET" required:"true"`
	UseScylla bool   `default:"false"        envconfig:"USE_SCYLLA"`
}

// LoadConfig loads the application config.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	if cfg.StreamURL == "" || cfg.JwtSecret == "" {
		return nil, fmt.Errorf("%w", errInvalidStreamURL)
	}

	return &cfg, nil
}
