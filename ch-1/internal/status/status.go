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

	"github.com/go-chi/chi/v5"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	wikimedia "github.com/codyonesock/backend_learning/ch-6/proto"
)

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

// Handler returns the router for /status routes.
func (s *Service) Handler(statusService *Service, streamURL string) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := statusService.ProcessStream(ctx, streamURL); err != nil {
			s.Logger.Error("Error processing stream", zap.Error(err))
			http.Error(w, "Error processing stream", http.StatusInternalServerError)
		}
	})

	return r
}

// ProcessStream reads the recent change stream and updates stats.
func (s *Service) ProcessStream(ctx context.Context, streamURL string) error {
	parsedURL, err := s.validateStreamURL(streamURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, s.ContextTimeout)
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

	processFunc := func(line string) error {
		return s.handleStreamData(line)
	}

	return streamReader(ctx, res.Body, processFunc)
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

// handleStreamData takes in lines of data from the stream to update the stats.
func (s *Service) handleStreamData(line string) error {
	jsonData := strings.TrimPrefix(line, "data:")
	jsonData = strings.TrimSpace(jsonData)

	var rc shared.RecentChange
	if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
		s.Logger.Error("Error parsing JSON", zap.Error(err))
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	s.StatsInterface.UpdateStats(rc)
	time.Sleep(s.SleepTime) // Spam annoying :(

	return nil
}

// Producer interface.
type Producer interface {
	Produce(ctx context.Context, record *kgo.Record, cb func(*kgo.Record, error))
}

// StreamAndProduce reads the wikimedia stream and produces each event to Redpanda.
func StreamAndProduce(ctx context.Context, streamURL string, producer Producer, logger *zap.Logger) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch stream from URL %s: %w", streamURL, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	processFunc := func(line string) error {
		var rc shared.RecentChange
		if err := json.Unmarshal([]byte(line[5:]), &rc); err != nil {
			logger.Warn("failed to unmarshal event", zap.Error(err))
			return nil
		}

		pb := &wikimedia.RecentChange{
			User:      rc.User,
			Bot:       rc.Bot,
			ServerUrl: rc.ServerURL,
		}

		eventBytes, err := proto.Marshal(pb)
		if err != nil {
			logger.Warn("failed to marshal event", zap.Error(err))
			return nil
		}

		record := &kgo.Record{
			Value: eventBytes,
		}

		producer.Produce(ctx, record, func(_ *kgo.Record, err error) {
			if err != nil {
				logger.Warn("failed to produce to Redpanda", zap.Error(err))
			}
		})

		return nil
	}

	return streamReader(ctx, resp.Body, processFunc)
}

// streamReader is a helper function that reads a stream and processes each line.
func streamReader(ctx context.Context, streamBody io.Reader, processFunc func(line string) error) error {
	br := bufio.NewReader(streamBody)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled or timed out: %w", ctx.Err())
		default:
			line, err := br.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil // End of stream
				}

				return fmt.Errorf("error reading line: %w", err)
			}

			if strings.HasPrefix(line, "data:") {
				if err := processFunc(line); err != nil {
					return fmt.Errorf("error processing stream data: %w", err)
				}
			}
		}
	}
}
