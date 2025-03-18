package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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

// GetStats handles /stats, it returns the updated StatsResponse data
func GetStats(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(string(statsJSON))
}
