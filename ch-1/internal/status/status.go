package status

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://stream.wikimedia.org/v2/stream/recentchange")
	if err != nil {
		log.Fatalf("Error getting recent change stream: %v", err)
		http.Error(w, "Error connecting to stream", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			log.Printf("Error reading body: %v", err)
			break
		}

		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimPrefix(line, "data: ")
			jsonData = strings.TrimSpace(jsonData)

			var rc RecentChange
			if err := json.Unmarshal([]byte(jsonData), &rc); err != nil {
				log.Printf("Error parsing: %v", err)
			} else {
				updateStats(rc)
			}

			// Spam annoying :(
			time.Sleep(5 * time.Second)
		}
	}
}

func updateStats(rc RecentChange) {
	stats.Mu.Lock()
	defer stats.Mu.Unlock()

	stats.WikiStats.MessagesConsumed++
	stats.WikiStats.DistinctUsers[rc.User]++
	stats.WikiStats.DistinctServerURLs[rc.ServerURL]++

	if rc.Bot {
		stats.WikiStats.BotsCount++
	} else {
		stats.WikiStats.NonBotsCount++
	}
}
