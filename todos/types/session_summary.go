package types

import "time"

// SessionSummary is the list projection of the active run's monitored thread.
type SessionSummary struct {
	Found      bool      `json:"found"`
	InProgress bool      `json:"inProgress"`
	State      string    `json:"state"`
	DurationMS int64     `json:"durationMs"`
	CostUSD    float64   `json:"costUsd"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
