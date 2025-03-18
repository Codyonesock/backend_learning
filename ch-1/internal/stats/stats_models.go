package stats

// Stats holds the core data that comes from Wikimedia
type Stats struct {
	MessagesConsumed   int            `json:"messages_consumed"`
	DistinctUsers      map[string]int `json:"-"`
	BotsCount          int            `json:"bots_count"`
	NonBotsCount       int            `json:"non_bots_count"`
	DistinctServerURLs map[string]int `json:"-"`
}

// StatsResponse returns all the counts in ints
type StatsResponse struct {
	MessagesConsumed       int `json:"messages_consumed"`
	DistinctUsersCount     int `json:"distinct_users"`
	BotsCount              int `json:"bots_count"`
	NonBotsCount           int `json:"non_bots_count"`
	DistinctServerURLCount int `json:"distinct_server_urls"`
}

// DistinctUsersCount is a method that returns the length of DistinctUsers
func (s *Stats) DistinctUsersCount() int {
	return len(s.DistinctUsers)
}

// DistinctServerURLCount is a method that returns the length of DistinctServerURLCount
func (s *Stats) DistinctServerURLCount() int {
	return len(s.DistinctServerURLs)
}
