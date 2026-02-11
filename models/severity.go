package models

import "strings"

// Severity represents the severity level of a change
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
	Info     Severity = "info"
)

// Value returns the numeric value for ordering/comparison
// Higher values indicate higher severity
func (s Severity) Value() int {
	switch s {
	case Critical:
		return 5
	case High:
		return 4
	case Medium:
		return 3
	case Low:
		return 2
	case Info:
		return 1
	default:
		return 0
	}
}

// String returns the string representation
func (s Severity) String() string {
	return string(s)
}

// ParseSeverity converts a string to a Severity type
func ParseSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "critical":
		return Critical
	case "high":
		return High
	case "medium":
		return Medium
	case "low":
		return Low
	case "info":
		return Info
	default:
		return Medium // default fallback
	}
}

// Max returns the higher severity between two severities
func Max(a, b Severity) Severity {
	if a.Value() > b.Value() {
		return a
	}
	return b
}

// MaxSeverities returns the maximum severity from a slice
func MaxSeverities(severities []Severity) Severity {
	if len(severities) == 0 {
		return Medium
	}

	max := severities[0]
	for _, s := range severities[1:] {
		if s.Value() > max.Value() {
			max = s
		}
	}
	return max
}

// SeverityDistribution tracks counts of each severity level
type SeverityDistribution struct {
	Critical int `json:"critical,omitempty"`
	High     int `json:"high,omitempty"`
	Medium   int `json:"medium,omitempty"`
	Low      int `json:"low,omitempty"`
	Info     int `json:"info,omitempty"`
}

// Add increments the count for the given severity
func (sd *SeverityDistribution) Add(s Severity) {
	switch s {
	case Critical:
		sd.Critical++
	case High:
		sd.High++
	case Medium:
		sd.Medium++
	case Low:
		sd.Low++
	case Info:
		sd.Info++
	}
}

// Total returns the total count across all severities
func (sd *SeverityDistribution) Total() int {
	return sd.Critical + sd.High + sd.Medium + sd.Low + sd.Info
}

// Max returns the highest severity with non-zero count
func (sd *SeverityDistribution) Max() Severity {
	if sd.Critical > 0 {
		return Critical
	}
	if sd.High > 0 {
		return High
	}
	if sd.Medium > 0 {
		return Medium
	}
	if sd.Low > 0 {
		return Low
	}
	if sd.Info > 0 {
		return Info
	}
	return Medium
}
