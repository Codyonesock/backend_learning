// Package users handles login.
package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// User represents a user in the system.
type User struct {
	Username string
	Password string
}

// Service manages user sync.
type Service struct {
	logger *zap.Logger
	mu     sync.Mutex
	users  map[string]User
}

var errUserAlreadyExists = errors.New("user already exists")

// NewUserService creates a new user service.
func NewUserService(l *zap.Logger) *Service {
	return &Service{
		logger: l,
		mu:     sync.Mutex{},
		users:  make(map[string]User),
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

	s.logger.Info("User logged in successfully", zap.String("username", req.Username))
	w.WriteHeader(http.StatusOK)
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
