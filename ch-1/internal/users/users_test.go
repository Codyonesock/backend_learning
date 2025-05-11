package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/backend_learning/ch-1/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

func TestRegisterHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful registration", func(t *testing.T) {
		t.Parallel()

		l := zap.NewNop()
		s := users.NewUserService(l, "super-secure-random-key", 1*time.Second)

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
		s := users.NewUserService(l, "super-secure-random-key", 1*time.Second)

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
	s := users.NewUserService(l, "super-secure-random-key", 1*time.Hour)

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

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	l := zap.NewNop()
	s := users.NewUserService(l, "super-secure-random-key", 1*time.Second)

	handler := s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("valid token", func(t *testing.T) {
		t.Parallel()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "blub",
			"exp":      time.Now().Add(30 * time.Second).Unix(),
		})

		tokenString, err := token.SignedString([]byte("super-secure-random-key"))
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
