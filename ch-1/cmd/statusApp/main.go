package main

import (
	"fmt"
	"net/http"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
	"github.com/codyonesock/backend_learning/ch-1/internal/status"
)

func main() {
	//# curl http://localhost:7000/status
	http.HandleFunc("/status", status.GetStatus)

	//# curl http://localhost:7000/stats
	http.HandleFunc("/stats", stats.GetStats)

	port := ":7000"
	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(port, nil)
}

// * 1. Create a  basic Go application that listens on port 7000 and has a status endpoint
// * 	a. Its common for Go to put binaries under a cmd directory
// * 2. Create a process to consume the wikipedia recent changes stream https://stream.wikimedia.org/v2/stream/recentchange and log these to stdout
// * 3. Replace the logs, with an in-memory /stats endpoint that a user can hit to get the latest stats on what we’ve processed
// * 4. Create the following stats
// * 	a. Number of messages consumed
// * 	b. Number of distinct users
// * 	c. Number of bots & Number of non-bots
// * 	d. Count by distinct server URLs
// ! 5. Create tests for your application (if you didn’t already)
// ! 6. Run tests with the race detector on (-race)
