// Package githubpush opens a GitHub issue from a native TODO and records the
// resulting reference back on the TODO. It is one-way: nothing is read back
// from GitHub, and an already-linked TODO is refused unless the caller forces
// a second issue.
package githubpush

import (
	"fmt"
	"net/url"
	"strings"
)

// ResolveBaseURL picks the first non-empty candidate, most-specific first, and
// validates it is an absolute http(s) origin with the trailing slash removed.
// No candidate resolves to an empty string rather than an error: whether a base
// URL is actually required depends on the body being pushed.
func ResolveBaseURL(candidates ...string) (string, error) {
	for _, candidate := range candidates {
		candidate = strings.TrimRight(strings.TrimSpace(candidate), "/")
		if candidate == "" {
			continue
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "", fmt.Errorf("invalid base URL %q: expected an absolute http(s) origin like https://gavel.example.com", candidate)
		}
		return candidate, nil
	}
	return "", nil
}

// IsLoopback reports whether a resolved base URL points at the local machine,
// which means anything rendering the body elsewhere cannot fetch its images.
func IsLoopback(baseURL string) bool {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	switch parsed.Hostname() {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0":
		return true
	default:
		return false
	}
}
