package prwatch

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/github"
)

var severityBadgeRegex = regexp.MustCompile(`_(?:🔴\s*Critical|🟠\s*Major|🟡\s*Minor)_`)

func parseSeverityFromBadge(body string) string {
	m := severityBadgeRegex.FindString(body)
	switch {
	case strings.Contains(m, "Critical"):
		return "critical"
	case strings.Contains(m, "Major"):
		return "major"
	case strings.Contains(m, "Minor"):
		return "minor"
	default:
		return ""
	}
}

// Matches file path with optional count suffix: "pkg/handler.go (2)" or "pkg/handler.go"
var nitpickFilePathRegex = regexp.MustCompile(`^([^\s(]+\.\w+)`)

// Matches line references like `44-107`: or `10`:
var nitpickLineRegex = regexp.MustCompile("^`(\\d+)(?:-\\d+)?`:")

func parseNitpickComments(comment github.PRComment) []github.PRComment {
	outerBody := extractNestedDetailsBody(comment.Body, "Nitpick comments")
	if outerBody == "" {
		return nil
	}

	fileBlocks := parseNestedFileBlocks(outerBody)
	var results []github.PRComment
	for _, fb := range fileBlocks {
		if fb.Body == "" {
			continue
		}
		results = append(results, github.PRComment{
			ID:       comment.ID,
			Author:   comment.Author,
			URL:      comment.URL,
			Path:     fb.Path,
			Line:     fb.Line,
			Body:     fb.Body,
			Severity: "nitpick",
		})
	}
	return results
}

type nitpickFileBlock struct {
	Path string
	Line int
	Body string
}

func parseNestedFileBlocks(html string) []nitpickFileBlock {
	var results []nitpickFileBlock
	idx := 0
	for {
		start := strings.Index(html[idx:], "<details>")
		if start == -1 {
			break
		}
		start += idx

		sumStart := strings.Index(html[start:], "<summary>")
		if sumStart == -1 {
			break
		}
		sumStart += start + len("<summary>")
		sumEnd := strings.Index(html[sumStart:], "</summary>")
		if sumEnd == -1 {
			break
		}
		summary := strings.TrimSpace(html[sumStart : sumStart+sumEnd])

		// Extract file path from summary like "formatters/html_formatter.go (2)"
		path := ""
		if m := nitpickFilePathRegex.FindStringSubmatch(summary); len(m) > 1 {
			path = m[1]
		}

		// Find the matching </details> respecting nesting
		bodyStart := sumStart + sumEnd + len("</summary>")
		body := extractNestedBody(html, bodyStart)

		// Strip <blockquote> wrapper
		body = strings.TrimSpace(body)
		body = strings.TrimPrefix(body, "<blockquote>")
		body = strings.TrimSuffix(body, "</blockquote>")
		body = strings.TrimSpace(body)

		// Strip nested <details> from body (suggested fix blocks etc)
		body = stripNestedDetails(body)
		body = strings.TrimSpace(body)

		// Extract line number from body like `44-107`:
		line := 0
		if m := nitpickLineRegex.FindStringSubmatch(body); len(m) > 1 {
			line, _ = strconv.Atoi(m[1])
		}

		// Also strip "Also applies to:" lines
		if applyIdx := strings.Index(body, "Also applies to:"); applyIdx != -1 {
			body = strings.TrimSpace(body[:applyIdx])
		}

		if body != "" {
			results = append(results, nitpickFileBlock{Path: path, Line: line, Body: body})
		}

		// Advance past this details block
		closeIdx := findMatchingClose(html, start)
		if closeIdx == -1 {
			break
		}
		idx = closeIdx
	}
	return results
}

func extractNestedBody(html string, bodyStart int) string {
	depth := 1
	pos := bodyStart
	for depth > 0 && pos < len(html) {
		nextOpen := strings.Index(html[pos:], "<details>")
		nextClose := strings.Index(html[pos:], "</details>")
		if nextClose == -1 {
			break
		}
		if nextOpen != -1 && nextOpen < nextClose {
			depth++
			pos += nextOpen + len("<details>")
		} else {
			depth--
			if depth == 0 {
				return strings.TrimSpace(html[bodyStart : pos+nextClose])
			}
			pos += nextClose + len("</details>")
		}
	}
	return ""
}

func findMatchingClose(html string, start int) int {
	depth := 0
	pos := start
	for pos < len(html) {
		nextOpen := strings.Index(html[pos:], "<details>")
		nextClose := strings.Index(html[pos:], "</details>")
		if nextClose == -1 {
			return -1
		}
		if nextOpen != -1 && nextOpen < nextClose {
			depth++
			pos += nextOpen + len("<details>")
		} else {
			depth--
			if depth == 0 {
				return pos + nextClose + len("</details>")
			}
			pos += nextClose + len("</details>")
		}
	}
	return -1
}

var nestedDetailsRegex = regexp.MustCompile(`(?s)<details>.*?</details>`)

func stripNestedDetails(body string) string {
	return nestedDetailsRegex.ReplaceAllString(body, "")
}

func extractNestedDetailsBody(html, summaryPrefix string) string {
	idx := 0
	for {
		start := strings.Index(html[idx:], "<details>")
		if start == -1 {
			return ""
		}
		start += idx

		sumStart := strings.Index(html[start:], "<summary>")
		if sumStart == -1 {
			return ""
		}
		sumStart += start + len("<summary>")
		sumEnd := strings.Index(html[sumStart:], "</summary>")
		if sumEnd == -1 {
			return ""
		}
		summary := strings.TrimSpace(html[sumStart : sumStart+sumEnd])
		stripped := leadingNonASCII.ReplaceAllString(summary, "")

		if !strings.HasPrefix(stripped, summaryPrefix) {
			idx = start + len("<details>")
			continue
		}

		bodyStart := sumStart + sumEnd + len("</summary>")
		body := extractNestedBody(html, bodyStart)
		if body != "" {
			// Strip outer blockquote wrapper
			body = strings.TrimSpace(body)
			body = strings.TrimPrefix(body, "<blockquote>")
			body = strings.TrimSuffix(body, "</blockquote>")
			return strings.TrimSpace(body)
		}
		return ""
	}
}
