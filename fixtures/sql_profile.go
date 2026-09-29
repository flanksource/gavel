package fixtures

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type SQLProfile struct {
	Path            string                `json:"path"`
	QueryCount      int                   `json:"query_count"`
	SlowQueryCount  int                   `json:"slow_query_count"`
	TotalDurationMS float64               `json:"total_duration_ms"`
	MaxQueryMS      float64               `json:"max_query_ms"`
	Statements      []SQLProfileStatement `json:"statements,omitempty"`
}

type SQLProfileStatement struct {
	SQL        string   `json:"sql"`
	Params     []string `json:"params"`
	DurationMS float64  `json:"duration_ms"`
	Rows       int64    `json:"rows"`
	Slow       bool     `json:"slow"`
	Error      bool     `json:"error"`
}

type sqlProfileEvent struct {
	DurationNS *int64   `json:"duration_ns"`
	SQL        string   `json:"sql"`
	Params     []string `json:"params"`
	Rows       int64    `json:"rows"`
	Slow       bool     `json:"slow"`
	Error      bool     `json:"error"`
}

func ReadSQLProfile(path string) (*SQLProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open SQL profile %s: %w", path, err)
	}
	defer file.Close()
	profile := &SQLProfile{Path: path}
	decoder := json.NewDecoder(file)
	for {
		var event sqlProfileEvent
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode SQL profile %s at line %d: %w", path, profile.QueryCount+1, err)
		}
		if event.DurationNS == nil || *event.DurationNS < 0 {
			return nil, fmt.Errorf("SQL profile %s has missing or negative duration at line %d", path, profile.QueryCount+1)
		}
		profile.QueryCount++
		if event.Slow {
			profile.SlowQueryCount++
		}
		ms := float64(*event.DurationNS) / 1e6
		profile.TotalDurationMS += ms
		profile.MaxQueryMS = max(profile.MaxQueryMS, ms)
		if event.SQL != "" {
			if event.Params == nil {
				event.Params = []string{}
			}
			profile.Statements = append(profile.Statements, SQLProfileStatement{
				SQL: event.SQL, Params: event.Params, DurationMS: ms,
				Rows: event.Rows, Slow: event.Slow, Error: event.Error,
			})
		}
	}
	return profile, nil
}
