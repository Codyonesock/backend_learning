// Package logger is to configure level logging
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// CreateLogger initializes a logger with the specified log level.
func CreateLogger(logLevel string) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()

	switch logLevel {
	case "DEBUG":
		level.SetLevel(zap.DebugLevel)
	case "INFO":
		level.SetLevel(zap.InfoLevel)
	case "ERROR":
		level.SetLevel(zap.ErrorLevel)
	case "PANIC":
		level.SetLevel(zap.PanicLevel)
	default:
		level.SetLevel(zap.InfoLevel)
	}

	cfg := zap.Config{
		Level:            level,
		Development:      logLevel == "DEBUG",
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
