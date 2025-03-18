package status

// RecentChange is a based on event data from Wikimedia
type RecentChange struct {
	User      string `json:"user"`
	Bot       bool   `json:"bot"`
	ServerURL string `json:"server_url"`
}
