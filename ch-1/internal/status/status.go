package status

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/codyonesock/backend_learning/ch-1/internal/stats"
)

// GetStatus handles /status, it reads recentchange updates from Wikimedia
// It processes events from the stream and updates stats
func GetStatus(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://stream.wikimedia.org/v2/stream/recentchange")
	if err != nil {
		log.Printf("Error getting recent change stream: %v", err)
		http.Error(w, "Error connecting to stream", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	br := bufio.NewReader(res.Body)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			// Idk if it ever ends, but if it does, exit :D
			if err == io.EOF {
				log.Println("Stream ended")
				break
			}
			log.Printf("Error reading body: %v", err)
			break
		}

		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimPrefix(line, "data:")
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

// updateStats takes in a RecentChange struct,
// which in return updates WikiStats
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
