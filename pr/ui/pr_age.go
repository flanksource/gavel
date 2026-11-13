package ui

import (
	"strings"
	"time"
)

func prAgeBucket(updated, now time.Time) string {
	if updated.IsZero() {
		return "unknown"
	}
	days := now.Sub(updated).Hours() / 24
	switch {
	case days < 1:
		return "day"
	case days < 7:
		return "week"
	case days < 30:
		return "month"
	default:
		return "older"
	}
}

func matchesPRAge(bucket string, filters []string) bool {
	hasIncludes, included := false, false
	for _, filter := range filters {
		if excluded, ok := strings.CutPrefix(filter, "-"); ok {
			if excluded == bucket {
				return false
			}
		} else {
			hasIncludes = true
			included = included || filter == bucket
		}
	}
	return !hasIncludes || included
}
