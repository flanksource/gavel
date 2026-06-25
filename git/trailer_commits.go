package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/models"
)

// maxCommitDiffBytes caps the size of a single commit's rendered diff so an
// enormous commit cannot flood the dashboard; the remainder is truncated.
const maxCommitDiffBytes = 256 * 1024

var commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{4,64}$`)

// IsValidCommitHash reports whether s is a syntactically valid abbreviated or
// full git object hash, so callers can reject untrusted input before shelling
// out to git.
func IsValidCommitHash(s string) bool {
	return commitHashPattern.MatchString(strings.TrimSpace(s))
}

// CommitsWithTrailer returns the commits reachable from any ref whose git
// trailer `key` equals `value`, newest-first. It pre-filters the log with a
// fixed-string `--grep` on the "Key: value" trailer line, then confirms each
// candidate against the parsed trailers so a coincidental body line that merely
// mentions the phrase never counts. An empty key/value yields no commits.
func CommitsWithTrailer(path, key, value string) (models.Commits, error) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return nil, nil
	}

	grep := key + ": " + value
	cmd := exec.Command("git", "log", "--all", "--date=iso-strict",
		"--fixed-strings", "--grep="+grep, commitLogPrettyFormat)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// A repository with no commits yet has no trailers to match; that is a
		// valid empty result, not a failure.
		if isNoCommitsError(output) {
			return nil, nil
		}
		return nil, fmt.Errorf("git log --grep %q: %w\nOutput: %s", grep, err, string(output))
	}

	commits, err := ParseGitLogOutput(output)
	if err != nil {
		return nil, err
	}
	matched := make(models.Commits, 0, len(commits))
	for _, c := range commits {
		if strings.TrimSpace(c.Trailers[key]) == value {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

// isNoCommitsError reports whether git log failed only because the repository
// has no commits / refs yet (as opposed to a real error).
func isNoCommitsError(output []byte) bool {
	s := string(output)
	return strings.Contains(s, "does not have any commits yet") ||
		strings.Contains(s, "bad default revision") ||
		strings.Contains(s, "unknown revision")
}

// CommitDiff returns the colored `git show` output (diffstat + patch) for a
// single commit, so the dashboard can render it through an ANSI viewer. Output
// is capped at maxCommitDiffBytes; the bool reports whether it was truncated.
func CommitDiff(path, hash string) (string, bool, error) {
	hash = strings.TrimSpace(hash)
	if !IsValidCommitHash(hash) {
		return "", false, fmt.Errorf("invalid commit hash %q", hash)
	}
	cmd := exec.Command("git", "show", "--color=always", "--stat", "--patch", hash)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", false, fmt.Errorf("git show %s: %w\nOutput: %s", hash, err, string(out))
	}
	if len(out) > maxCommitDiffBytes {
		return truncateAtLine(string(out), maxCommitDiffBytes) +
			"\n\n… diff truncated (showing first 256 KB) …\n", true, nil
	}
	return string(out), false, nil
}

// truncateAtLine trims s to at most max bytes, backing up to the last newline so
// it never cuts a line (and rarely an ANSI escape) mid-sequence.
func truncateAtLine(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	if nl := strings.LastIndexByte(cut, '\n'); nl > 0 {
		return cut[:nl]
	}
	return cut
}

// RemoteWebURL returns the https web base for the repository's origin remote
// (e.g. https://github.com/owner/repo), or "" when there is no origin or it is
// not a recognizable host/owner/repo URL. Callers append "/commit/<hash>" to
// build a commit link; a local-only repo simply has no link.
func RemoteWebURL(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin in %s: %w", path, err)
	}
	return remoteToWebURL(strings.TrimSpace(string(out))), nil
}

// remoteToWebURL converts a git remote URL into its https web base, handling the
// scp-like ssh form (git@host:owner/repo.git) and scheme URLs
// (ssh://, https://, http://). Returns "" for anything it cannot parse into a
// host + owner/repo path.
func remoteToWebURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	remote = strings.TrimSuffix(remote, ".git")

	var host, repoPath string
	switch {
	case strings.HasPrefix(remote, "git@"):
		host, repoPath, _ = strings.Cut(strings.TrimPrefix(remote, "git@"), ":")
	case strings.Contains(remote, "://"):
		_, rest, _ := strings.Cut(remote, "://")
		// Strip userinfo (user@) that precedes the host.
		if at := strings.Index(rest, "@"); at >= 0 {
			if slash := strings.Index(rest, "/"); slash < 0 || at < slash {
				rest = rest[at+1:]
			}
		}
		host, repoPath, _ = strings.Cut(rest, "/")
	default:
		return ""
	}

	host = strings.TrimSpace(host)
	repoPath = strings.Trim(strings.TrimSpace(repoPath), "/")
	if host == "" || repoPath == "" {
		return ""
	}
	return "https://" + host + "/" + repoPath
}
