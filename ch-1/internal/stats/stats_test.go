package stats_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestGetStats(t *testing.T) {
// 	WikiStats = Stats{
// 		MessagesConsumed:   10,
// 		DistinctUsers:      map[string]int{"user1": 1, "user2": 2},
// 		DistinctServerURLs: map[string]int{"https://blub.com": 3},
// 		BotsCount:          4,
// 		NonBotsCount:       6,
// 	}

// 	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
// 	recorder := httptest.NewRecorder()

// 	GetStats(recorder, req)

// 	if recorder.Code != http.StatusOK {
// 		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
// 	}

// 	// StatsResponse value
// 	expectedBody := `{"messages_consumed":10,"distinct_users":2,"bots_count":4,"non_bots_count":6,"distinct_server_urls":1}`
// 	if recorder.Body.String() != expectedBody {
// 		t.Errorf("expected body %s, got %s", expectedBody, recorder.Body.String())
// 	}
// }
