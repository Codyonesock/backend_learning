package stats

import (
	"encoding/json"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

var (
	// Mu is a mutex that gets used by both /status and /stats to prevent race conditions
	Mu sync.Mutex

	// WikiStats is used to hold Stats data as it updates
	WikiStats = Stats{
		DistinctUsers:      map[string]int{},
		DistinctServerURLs: map[string]int{},
	}
)

func GetStats(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	Mu.Lock()
	defer Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	response := StatsResponse{
		MessagesConsumed:       WikiStats.MessagesConsumed,
		DistinctUsersCount:     WikiStats.DistinctUsersCount(),
		BotsCount:              WikiStats.BotsCount,
		NonBotsCount:           WikiStats.NonBotsCount,
		DistinctServerURLCount: WikiStats.DistinctServerURLCount(),
	}
	statsJSON, _ := json.Marshal(response)
	w.Write(statsJSON)
	logger.Info("Current stats", zap.Any("response", response))
}
