package status

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
)

// StatusServiceInterface defines methods for StatusService
type StatusServiceInterface interface {
	ProcessStream(streamURL string) error
	UpdateStats(rc RecentChange)
}

// StatusService handles dependencies
type StatusService struct {
	Logger *zap.Logger
}

// NewStatusService create a new instance of StatusService
func NewStatusService(l *zap.Logger) StatusServiceInterface {
	return &StatusService{Logger: l}
}

// RecentChange is based on event data from Wikimedia
type RecentChange struct {
	User      string `json:"user"`
	Bot       bool   `json:"bot"`
	ServerURL string `json:"server_url"`
}

// ProcessStream reads the recent change stream and updates stats
func (s *StatusService) ProcessStream(streamURL string) error {
	res, err := http.Get(streamURL)
	if err != nil {
		s.Logger.Error("Error getting recent change stream", zap.String("stream_url", streamURL), zap.Error(err))
		return err
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.Logger.Info("Stream ended")
				break
			}
			s.Logger.Error("Error reading body", zap.Error(err))
			return err
		}

		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimPrefix(line, "data:")
			jsonData = strings.TrimSpace(jsonData)

			var rc RecentChange
			if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
				s.Logger.Error("Error parsing JSON", zap.Error(err))
			} else {
				s.UpdateStats(rc)
			}

			// Spam annoying :(
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

// UpdateStats updates the WikiStats with the given RecentChange
func (s *StatusService) UpdateStats(rc RecentChange) {
	stats.Mu.Lock()
	defer stats.Mu.Unlock()

	stats.WikiStats.MessagesConsumed++
	stats.WikiStats.DistinctUsers[rc.User]++
	stats.WikiStats.DistinctServerURLs[rc.ServerURL]++

	if rc.Bot {
		stats.WikiStats.BotsCount++
	} else {
		stats.WikiStats.NonBotsCount++
	}

	s.Logger.Info("Stats updated", zap.String("user", rc.User), zap.Bool("bot", rc.Bot))
}
