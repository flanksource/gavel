package githubpush

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Ref names the issue a push writes to. A zero Number opens a new issue; an
// empty Repo falls back to the workspace's own origin remote.
type Ref struct {
	Repo   string
	Number int
}

// URL is the canonical github.com address for the reference. It is empty until
// both halves are known, since a repo-less or numberless reference has no page.
func (r Ref) URL() string {
	if r.Repo == "" || r.Number <= 0 {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/issues/%d", r.Repo, r.Number)
}

// issueRefPattern accepts every shape a user has at hand: a bare number, an
// `owner/repo#number` alias (what this package records), and the issue URL that
// GitHub puts in the address bar.
var issueRefPattern = regexp.MustCompile(
	`^(?:https?://[^/]+/)?(?:([\w.-]+/[\w.-]+)(?:/issues/|#)|#)?(\d+)$`)

// ParseIssueRef reads a GitHub issue reference in any of the shapes above. It
// is also how the aliases this package writes are read back.
func ParseIssueRef(ref string) (Ref, error) {
	match := issueRefPattern.FindStringSubmatch(strings.TrimSpace(ref))
	if match == nil {
		return Ref{}, fmt.Errorf("%q is not a GitHub issue reference: "+
			"use 123, owner/repo#123, or https://github.com/owner/repo/issues/123", ref)
	}
	number, err := strconv.Atoi(match[2])
	if err != nil || number <= 0 {
		return Ref{}, fmt.Errorf("%q is not a GitHub issue number", ref)
	}
	return Ref{Repo: match[1], Number: number}, nil
}
