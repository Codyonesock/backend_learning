// Package shared is for shared stuff.
package shared

// RecentChange is based on event data from Wikimedia.
type RecentChange struct {
	User      string `json:"user"`
	Bot       bool   `json:"bot"`
	ServerURL string `json:"server_url"`
}

// Stats holds the core data that comes from Wikimedia.
type Stats struct {
	MessagesConsumed   int            `json:"messages_consumed"`
	DistinctUsers      map[string]int `json:"-"`
	BotsCount          int            `json:"bots_count"`
	NonBotsCount       int            `json:"non_bots_count"`
	DistinctServerURLs map[string]int `json:"-"`
}
