package stats_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/shared"
	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
)

// MockStorage is a mock implementation of the shared.Storage interface.
type MockStorage struct {
	SaveStatsFunc func(*shared.Stats) error
	LoadStatsFunc func() (*shared.Stats, error)
	Stats         *shared.Stats
}

func (m *MockStorage) SaveStats(stat *shared.Stats) error {
	if m.SaveStatsFunc != nil {
		return m.SaveStatsFunc(stat)
	}

	m.Stats = stat

	return nil
}

func (m *MockStorage) LoadStats() (*shared.Stats, error) {
	if m.LoadStatsFunc != nil {
		return m.LoadStatsFunc()
	}

	if m.Stats == nil {
		m.Stats = &shared.Stats{
			MessagesConsumed:   0,
			DistinctUsers:      map[string]int{},
			BotsCount:          0,
			NonBotsCount:       0,
			DistinctServerURLs: map[string]int{},
		}
	}

	return m.Stats, nil
}

// Helper function to create a new stats.Service with a mock storage.
func newTestService(mockStorage *MockStorage) *stats.Service {
	logger := zap.NewNop()
	return stats.NewStatsService(logger, mockStorage)
}

// TestUpdateStats verifies that stats are updated correctly.
func TestUpdateStats(t *testing.T) {
	t.Parallel()

	mockStorage := &MockStorage{
		SaveStatsFunc: nil,
		LoadStatsFunc: nil,
		Stats: &shared.Stats{
			MessagesConsumed:   0,
			DistinctUsers:      map[string]int{},
			BotsCount:          0,
			NonBotsCount:       0,
			DistinctServerURLs: map[string]int{},
		},
	}
	service := newTestService(mockStorage)

	rc := shared.RecentChange{
		User:      "blub_user",
		Bot:       false,
		ServerURL: "https://blub.com",
	}

	service.UpdateStats(rc)

	service.Mu.Lock()
	defer service.Mu.Unlock()

	if service.Stats.MessagesConsumed != 1 {
		t.Errorf("expected MessagesConsumed to be 1, got %d", service.Stats.MessagesConsumed)
	}

	if service.Stats.DistinctUsers["blub_user"] != 1 {
		t.Errorf("expected DistinctUsers to be 1, got %d", service.Stats.DistinctUsers["blub_user"])
	}

	if service.Stats.BotsCount != 0 {
		t.Errorf("expected BotsCount to be 0, got %d", service.Stats.BotsCount)
	}

	if service.Stats.NonBotsCount != 1 {
		t.Errorf("expected NonBotsCount to be 1, got %d", service.Stats.NonBotsCount)
	}

	if service.Stats.DistinctServerURLs["https://blub.com"] != 1 {
		t.Errorf(
			"expected DistinctServerURLs to be 1, got %d",
			service.Stats.DistinctServerURLs["https://blub.com"],
		)
	}
}

// TestGetStats simulates a stats call annd verifies the response.
func TestGetStats(t *testing.T) {
	t.Parallel()

	mockStorage := &MockStorage{
		SaveStatsFunc: nil,
		LoadStatsFunc: nil,
		Stats: &shared.Stats{
			MessagesConsumed:   10,
			DistinctUsers:      map[string]int{"user1": 1, "user2": 2},
			BotsCount:          4,
			NonBotsCount:       6,
			DistinctServerURLs: map[string]int{"https://blub.com": 3},
		},
	}

	service := newTestService(mockStorage)

	if err := service.LoadStats(); err != nil {
		t.Fatalf("failed to load stats: %v", err)
	}

	recorder := httptest.NewRecorder()
	if err := service.GetStats(recorder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	expectedResponse := stats.Response{
		MessagesConsumed:       10,
		DistinctUsersCount:     2,
		BotsCount:              4,
		NonBotsCount:           6,
		DistinctServerURLCount: 1,
	}

	var actualResponse stats.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if expectedResponse != actualResponse {
		t.Errorf("expected response %+v, got %+v", expectedResponse, actualResponse)
	}
}
