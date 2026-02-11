package git

import (
	"regexp"
	"strings"

	. "github.com/flanksource/gavel/models"
)

var commitTypeScopeRegex = regexp.MustCompile(`^(\w+)(\(([^)]+)\))?:\s*(.+)$`)
var commitTypeRegex = regexp.MustCompile(`^(\w+)\s+:\s*(.+)$`)

// parses a conventional commit subject line (e.g. feat(scope): subject) into its type, scope, and subject components
func parseCommitTypeAndScope(subject string) (CommitType, ScopeType, string) {
	matches := commitTypeScopeRegex.FindStringSubmatch(subject)
	if len(matches) == 5 {
		commitType := CommitType(matches[1])
		scope := ScopeType(matches[3])
		subject := matches[4]
		return commitType, scope, subject
	}
	matches = commitTypeRegex.FindStringSubmatch(subject)
	if len(matches) == 3 {
		commitType := CommitType(matches[1])
		subject := matches[2]
		return commitType, ScopeTypeUnknown, subject
	}
	return CommitTypeUnknown, ScopeTypeUnknown, subject

}

func trim(s string) string {
	return strings.TrimSpace(s)
}

var refRegex = regexp.MustCompile(`#(\d+)`)
var refWithParansRegex = regexp.MustCompile(`\(#(\d+)\)`)

// parses a reference from the subject line (e.g. "subject (#1234)" -> "subject", "1234")
func parseReference(subject string) (string, string) {
	var ref string
	matches := refWithParansRegex.FindStringSubmatch(subject)
	if len(matches) == 2 {
		ref = matches[1]
		subject = strings.ReplaceAll(subject, matches[0], "")
	} else {
		matches = refRegex.FindStringSubmatch(subject)
		if len(matches) == 2 {
			ref = matches[1]
			subject = strings.ReplaceAll(subject, matches[0], "")
		}
	}

	return strings.TrimSpace(subject), strings.TrimSpace(ref)
}

var trailerKeys = []string{
	"Signed-off-by",
	"Co-authored-by",
	"Reviewed-by",
	"Acked-by",
	"Reported-by",
	"Tested-by",
	"Reviewed-by",
}

// find common trailers in commit messages like "Signed-off-by: ", "Co-authored-by: ", etc.
func parseTrailers(message string) (string, map[string]string) {
	trailers := make(map[string]string)
	lines := strings.Split(message, "\n")
	out := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, ":") {
			out += line + "\n"
			continue
		}
		for _, key := range trailerKeys {
			if strings.HasPrefix(line, key+":") {
				value := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
				trailers[key] = value
				goto NextLine
			}
		}
		out += line + "\n"
	NextLine:
	}
	return out, trailers
}
