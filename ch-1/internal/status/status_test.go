package status_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
// )

// func TestGetStatus(t *testing.T) {
// 	stats.Mu.Lock()
// 	stats.WikiStats = stats.Stats{
// 		DistinctUsers:      make(map[string]int),
// 		DistinctServerURLs: make(map[string]int),
// 	}
// 	stats.Mu.Unlock()

// 	recorder := httptest.NewRecorder()
// 	req := httptest.NewRequest(http.MethodGet, "/status", nil)

// 	go func() {
// 		GetStatus(recorder, req, "")
// 	}()

// 	if recorder.Code != http.StatusOK {
// 		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
// 	}

// 	t.Log("TestGetStatus completed")
// }

// func TestUpdateStats(t *testing.T) {
// 	stats.Mu.Lock()
// 	stats.WikiStats = stats.Stats{
// 		DistinctUsers:      make(map[string]int),
// 		DistinctServerURLs: make(map[string]int),
// 	}
// 	stats.Mu.Unlock()

// 	rc := RecentChange{
// 		User:      "blub",
// 		ServerURL: "https://blub.com",
// 		Bot:       false,
// 	}

// 	updateStats(rc)

// 	if stats.WikiStats.MessagesConsumed != 1 {
// 		t.Errorf("expected MessagesConsumed to be 1, got %d", stats.WikiStats.MessagesConsumed)
// 	}
// 	if stats.WikiStats.DistinctUsers["blub"] != 1 {
// 		t.Errorf("expected DistinctUsers[blub] to be 1, got %d", stats.WikiStats.DistinctUsers["blub"])
// 	}
// 	if stats.WikiStats.DistinctServerURLs["https://blub.com"] != 1 {
// 		t.Errorf("expected DistinctServerURLs[https://blub.com] to be 1, got %d", stats.WikiStats.DistinctServerURLs["https://blub.com"])
// 	}
// 	if stats.WikiStats.NonBotsCount != 1 {
// 		t.Errorf("expected NonBotsCount to be 1, got %d", stats.WikiStats.NonBotsCount)
// 	}
// 	if stats.WikiStats.BotsCount != 0 {
// 		t.Errorf("expected BotsCount to be 0, got %d", stats.WikiStats.BotsCount)
// 	}
// }
