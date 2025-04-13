// Package status reads and processes recent changes from the stream.
package status

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/models"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
)

// ServiceInterface defines methods for Service.
type ServiceInterface interface {
	ProcessStream(streamURL string) error
}

// Service handles dependencies and config.
type Service struct {
	Logger         *zap.Logger
	StatsInterface stats.ServiceInterface
	SleepTime      time.Duration
	ContextTimeout time.Duration
}

// NewStatusService create a new instance of Service.
func NewStatusService(l *zap.Logger, si stats.ServiceInterface, st, ct time.Duration) *Service {
	return &Service{
		Logger:         l,
		StatsInterface: si,
		SleepTime:      st,
		ContextTimeout: ct,
	}
}

// ProcessStream reads the recent change stream and updates stats.
func (s *Service) ProcessStream(streamURL string) error {
	parsedURL, err := s.validateStreamURL(streamURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	res, err := s.fetchStream(ctx, parsedURL)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			s.Logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	return s.processStreamData(res.Body)
}

// validateStreamURL will validate a url.
func (s *Service) validateStreamURL(streamURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(streamURL)
	if err != nil || !parsedURL.IsAbs() {
		s.Logger.Error("Invalid stream URL", zap.String("stream_url", streamURL), zap.Error(err))
		return nil, fmt.Errorf("invalid stream URL: %w", err)
	}

	return parsedURL, nil
}

// fetchStream will bind a parsedUrl to the context and return the response.
func (s *Service) fetchStream(ctx context.Context, parsedURL *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		s.Logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.Logger.Error("Error getting stream", zap.String("stream_url", parsedURL.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch stream from URL %s: %w", parsedURL.String(), err)
	}

	return res, nil
}

// processStreamData will attempt to process a stream body and pass the data to stats.
func (s *Service) processStreamData(body io.Reader) error {
	br := bufio.NewReader(body)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.Logger.Info("Stream ended")
				break
			}

			s.Logger.Error("Error reading line", zap.Error(err))

			return fmt.Errorf("error reading line: %w", err)
		}

		if strings.HasPrefix(line, "data:") {
			if err := s.handleStreamData(line); err != nil {
				s.Logger.Error("Error with stream", zap.Error(err))
				return fmt.Errorf("error with stream: %w", err)
			}
		}
	}

	return nil
}

// handleStreamData takes in lines of data from the stream to update the stats.
func (s *Service) handleStreamData(line string) error {
	jsonData := strings.TrimPrefix(line, "data:")
	jsonData = strings.TrimSpace(jsonData)

	var rc models.RecentChange
	if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
		s.Logger.Error("Error parsing JSON", zap.Error(err))
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	s.StatsInterface.UpdateStats(rc)
	time.Sleep(s.SleepTime) // Spam annoying :(

	return nil
}
