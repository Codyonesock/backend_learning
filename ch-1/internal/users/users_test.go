package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/users"
)

func TestRegisterHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful registration", func(t *testing.T) {
		t.Parallel()

		l := zap.NewNop()
		s := users.NewUserService(l)

		reqBody, _ := json.Marshal(map[string]string{
			"username": "blub",
			"password": "pw123",
		})
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		s.RegisterHandler(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("duplicate user", func(t *testing.T) {
		t.Parallel()

		l := zap.NewNop()
		s := users.NewUserService(l)

		if err := s.Register("blub", "pw123"); err != nil {
			t.Fatalf("unexpected error during setup: %v", err)
		}

		reqBody, _ := json.Marshal(map[string]string{
			"username": "blub",
			"password": "pw123",
		})
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		s.RegisterHandler(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}
	})
}

func TestLoginHandler(t *testing.T) {
	t.Parallel()

	l := zap.NewNop()
	s := users.NewUserService(l)

	if err := s.Register("blub", "pw123"); err != nil {
		t.Fatalf("unexpected error during setup: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		t.Parallel()

		reqBody, _ := json.Marshal(map[string]string{
			"username": "blub",
			"password": "pw123",
		})
		req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		s.LoginHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		t.Parallel()

		reqBody, _ := json.Marshal(map[string]string{
			"username": "blub",
			"password": "wrongpassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		s.LoginHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
