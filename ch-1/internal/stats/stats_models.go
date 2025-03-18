package stats

type Stats struct {
	MessagesConsumed   int            `json:"messages_consumed"`
	DistinctUsers      map[string]int `json:"-"`
	BotsCount          int            `json:"bots_count"`
	NonBotsCount       int            `json:"non_bots_count"`
	DistinctServerURLs map[string]int `json:"-"`
}

type StatsResponse struct {
	MessagesConsumed       int `json:"messages_consumed"`
	DistinctUsersCount     int `json:"distinct_users"`
	BotsCount              int `json:"bots_count"`
	NonBotsCount           int `json:"non_bots_count"`
	DistinctServerURLCount int `json:"distinct_server_urls"`
}

func (s *Stats) DistinctUsersCount() int {
	return len(s.DistinctUsers)
}

func (s *Stats) DistinctServerURLCount() int {
	return len(s.DistinctServerURLs)
}
