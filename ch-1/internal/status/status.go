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

// GetStatus handles /status, it reads recentchange updates from  Wikimedia
// It processes events from the stream and updates stats
func GetStatus(w http.ResponseWriter, r *http.Request, streamURL string, logger *zap.Logger) {
	res, err := http.Get(streamURL)
	if err != nil {
		logger.Error("Error getting recent change stream", zap.String("stream_url", streamURL), zap.Error(err))
		http.Error(w, "Error connecting to stream", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Stream ended")
				break
			}
			logger.Error("Error reading body", zap.Error(err))
			break
		}

		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimPrefix(line, "data:")
			jsonData = strings.TrimSpace(jsonData)

			var rc RecentChange
			if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
				logger.Error("Error parsing JSON", zap.Error(err))
			} else {
				updateStats(rc, logger)
			}

			// Spam annoying :(
			time.Sleep(5 * time.Second)
		}
	}
}

// updateStats takes in a RecentChange struct,
// which in return updates WikiStats
func updateStats(rc RecentChange, logger *zap.Logger) {
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

	logger.Info("Stats updated", zap.String("user", rc.User), zap.Bool("bot", rc.Bot))
}
