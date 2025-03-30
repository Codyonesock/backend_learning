package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

type Config struct {
	Port      string `json:"port"`
	StreamURL string `json:"stream_url"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return &config, nil
}

func main() {
	config, err := LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	//# curl http://localhost:7000/status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status.GetStatus(w, r, config.StreamURL)
	})

	//# curl http://localhost:7000/stats
	http.HandleFunc("/stats", stats.GetStats)

	fmt.Printf("Server running on port %s\n", config.Port)
	http.ListenAndServe(config.Port, nil)
}

//* 0. Finish setting up an initial dockerfile
//! 1. Structured Logging (Zap, zerolog, log, fast)
//! 2. Service pattern (receiver methods)
//! 3. Split out into proper entities (Getting rid of modules)
//! 4. Storage package agnostic/abstracted (Plug and play DB)
//! 5. Evaluate Golang routers (chi, mux, httprouter)
//! 6. Fix the tests :D
