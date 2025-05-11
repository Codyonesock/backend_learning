// Package users handles login.
package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// User represents a user in the system.
type User struct {
	Username string
	Password string
}

// Service manages user sync.
type Service struct {
	logger              *zap.Logger
	mu                  sync.Mutex
	users               map[string]User
	jwtSecret           string
	authTokenExpiration time.Duration
}

var errUserAlreadyExists = errors.New("user already exists")

// NewUserService creates a new user service.
func NewUserService(
	l *zap.Logger,
	jwt string,
	authExp time.Duration,
) *Service {
	return &Service{
		logger:              l,
		mu:                  sync.Mutex{},
		users:               make(map[string]User),
		jwtSecret:           jwt,
		authTokenExpiration: authExp,
	}
}

// RegisterHandler handles user registration.
func (s *Service) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode registration request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	if err := s.Register(req.Username, req.Password); err != nil {
		if errors.Is(err, errUserAlreadyExists) {
			s.logger.Error("User already exists", zap.String("username", req.Username))
			http.Error(w, "User already exists", http.StatusConflict)
		} else {
			s.logger.Error("Failed to register user", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	s.logger.Info("User registered successfully", zap.String("username", req.Username))
	w.WriteHeader(http.StatusCreated)
}

// LoginHandler handles user login.
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode login request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	if !s.Authenticate(req.Username, req.Password) {
		s.logger.Warn("Invalid login attempt", zap.String("username", req.Username))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)

		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(s.authTokenExpiration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	s.logger.Info("User logged in successfully", zap.String("username", req.Username))
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write([]byte(`{"token":"` + tokenString + `"}`)); err != nil {
		s.logger.Error("Failed to write auth token", zap.Error(err))
	}
}

// Register adds a new user to the service.
func (s *Service) Register(username, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return errUserAlreadyExists
	}

	s.users[username] = User{Username: username, Password: password}

	return nil
}

// Authenticate checks if the username and password are valid.
func (s *Service) Authenticate(username, password string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]

	return exists && user.Password == password
}

// AuthMiddleware validates the JWT in the auth header.
func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.logger.Warn("Missing or invalid Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(s.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			s.logger.Warn("Invalid or expired token", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
