package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ReviewScope struct {
	Type        string   // "diff", "range", "commit", "branch", "files", "pr", "date-range", "commits"
	CommitRange string   // for "range"
	Commit      string   // for "commit"
	Commits     []string // for "commits" (an explicit SHA set, e.g. an issue's commits)
	Branch      string   // for "branch"
	PRNumber    int      // for "pr"
	Files       []string // for "files"
	Since       string   // for "date-range"
	Until       string   // for "date-range"
}

func (s ReviewScope) String() string {
	switch s.Type {
	case "range":
		return fmt.Sprintf("commits %s", s.CommitRange)
	case "commit":
		return fmt.Sprintf("commit %s", s.Commit)
	case "commits":
		return fmt.Sprintf("commits [%s]", strings.Join(s.Commits, ", "))
	case "branch":
		return fmt.Sprintf("branch %s vs HEAD", s.Branch)
	case "pr":
		return fmt.Sprintf("PR #%d", s.PRNumber)
	case "date-range":
		return fmt.Sprintf("commits %s..%s", s.Since, s.Until)
	case "files":
		return fmt.Sprintf("files [%s]", strings.Join(s.Files, ", "))
	default:
		return "uncommitted diff"
	}
}

// IssueComment is one comment from the issue history, included in an issue-aware
// verification so the reviewer sees the full conversation, not just the body.
type IssueComment struct {
	Author string `json:"author,omitempty"`
	Body   string `json:"body,omitempty"`
}

// IssueContext makes a verification issue-aware: the reviewer scores the commits
// against the issue's spec and its stored acceptance criteria, rather than
// reviewing an uncommitted diff generically. CheckIDs are the static
// verify.AllChecks the issue selected as applicable; Criteria are the custom,
// functionality-specific criteria to score in addition.
type IssueContext struct {
	ID          string
	Title       string
	Description string
	SessionID   string
	Comments    []IssueComment
	Criteria    []string // custom acceptance criteria to score (one verdict each)
	CheckIDs    []string // selected static check IDs to enable
	CommitSHAs  []string // commits implementing the issue
}

var (
	prURLPattern  = regexp.MustCompile(`github\.com/.+/pull/(\d+)`)
	prHashPattern = regexp.MustCompile(`^#(\d+)$`)
	bareDigits    = regexp.MustCompile(`^\d{1,5}$`)
	dateRangeRe   = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\.\.(\d{4}-\d{2}-\d{2})$`)
	hexSHA        = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	refOffsetRe   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._/-]*[~^]\d*$`)
	headOffsetRe  = regexp.MustCompile(`^HEAD[~^]\d*$`)
)

// ClassifyArg returns the detected type and a normalized value for a single positional argument.
func ClassifyArg(arg, repoPath string) (argType string, value string) {
	if m := prURLPattern.FindStringSubmatch(arg); len(m) >= 2 {
		return "pr", m[1]
	}
	if m := prHashPattern.FindStringSubmatch(arg); len(m) >= 2 {
		return "pr", m[1]
	}
	if bareDigits.MatchString(arg) {
		return "pr", arg
	}
	if m := dateRangeRe.FindStringSubmatch(arg); len(m) == 3 {
		return "date-range", arg
	}
	if strings.Contains(arg, "..") {
		return "range", arg
	}
	if strings.ContainsAny(arg, "*?") {
		return "file", arg
	}
	if repoPath != "" {
		abs := arg
		if !filepath.IsAbs(arg) {
			abs = filepath.Join(repoPath, arg)
		}
		if info, err := os.Stat(abs); err == nil {
			if info.IsDir() {
				return "directory", arg
			}
			return "file", arg
		}
	}
	if strings.Contains(arg, "/") {
		return "file", arg
	}
	if hexSHA.MatchString(arg) {
		return "commit", arg
	}
	if headOffsetRe.MatchString(arg) || refOffsetRe.MatchString(arg) {
		return "commit", arg
	}
	return "branch", arg
}

func ResolveScope(args []string, commitRange, repoPath string) (ReviewScope, error) {
	if commitRange != "" {
		return ReviewScope{Type: "range", CommitRange: commitRange}, nil
	}
	if len(args) == 0 {
		return ReviewScope{Type: "diff"}, nil
	}

	var files []string
	var singular *ReviewScope

	for _, arg := range args {
		typ, val := ClassifyArg(arg, repoPath)
		switch typ {
		case "file", "directory":
			files = append(files, val)
		case "pr":
			n, _ := strconv.Atoi(val)
			s := ReviewScope{Type: "pr", PRNumber: n}
			if err := setSingular(&singular, s, arg); err != nil {
				return ReviewScope{}, err
			}
		case "range":
			s := ReviewScope{Type: "range", CommitRange: val}
			if err := setSingular(&singular, s, arg); err != nil {
				return ReviewScope{}, err
			}
		case "commit":
			s := ReviewScope{Type: "commit", Commit: val}
			if err := setSingular(&singular, s, arg); err != nil {
				return ReviewScope{}, err
			}
		case "branch":
			s := ReviewScope{Type: "branch", Branch: val}
			if err := setSingular(&singular, s, arg); err != nil {
				return ReviewScope{}, err
			}
		case "date-range":
			m := dateRangeRe.FindStringSubmatch(val)
			s := ReviewScope{Type: "date-range", Since: m[1], Until: m[2]}
			if err := setSingular(&singular, s, arg); err != nil {
				return ReviewScope{}, err
			}
		}
	}

	if singular != nil && len(files) > 0 {
		return ReviewScope{}, fmt.Errorf("cannot mix %s argument with file paths", singular.Type)
	}
	if singular != nil {
		return *singular, nil
	}
	return ReviewScope{Type: "files", Files: files}, nil
}

func setSingular(current **ReviewScope, next ReviewScope, arg string) error {
	if *current != nil {
		return fmt.Errorf("cannot combine %s (%s) with previous %s argument", next.Type, arg, (*current).Type)
	}
	*current = &next
	return nil
}
