package commit

import "strings"

// tokenLimitMarkers are case-insensitive substrings that AI providers use to
// signal that a prompt exceeded the model's context window. Errors are not
// typed across the clicky/commons-db/captain providers, so we match on the
// rendered message. Keep these lowercase.
var tokenLimitMarkers = []string{
	"context length",
	"context_length",
	"context window",
	"context limit",
	"prompt is too long",
	"too many tokens",
	"maximum context",
	"token limit",
	"exceeds the maximum",
	"input length and `max_tokens`",
}

// isTokenLimitError reports whether err indicates the AI prompt overflowed the
// model context window, so the caller can retry with a smaller, chunked diff.
func isTokenLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range tokenLimitMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
