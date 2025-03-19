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
