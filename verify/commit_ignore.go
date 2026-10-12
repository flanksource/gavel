package verify

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type CommitIgnoreMatch struct {
	File    string
	Pattern string
}

// MatchCommitIgnoreFiles applies commit.gitignore and commit.allow to supplied paths.
func MatchCommitIgnoreFiles(files, patterns, allow []string) ([]CommitIgnoreMatch, error) {
	if len(files) == 0 || len(patterns) == 0 {
		return nil, nil
	}

	blockers, err := parseCommitIgnorePatterns(patterns, "commit.gitignore")
	if err != nil {
		return nil, err
	}
	allowMatchers, err := parseCommitIgnorePatterns(allow, "commit.allow")
	if err != nil {
		return nil, err
	}
	blockMatcher := gitignore.NewMatcher(blockers)
	allowMatcher := gitignore.NewMatcher(allowMatchers)

	var matches []CommitIgnoreMatch
	for _, file := range files {
		parts := strings.Split(filepath.ToSlash(file), "/")
		if allowMatcher.Match(parts, false) || !blockMatcher.Match(parts, false) {
			continue
		}
		for i, pattern := range blockers {
			if pattern != nil && gitignore.NewMatcher([]gitignore.Pattern{pattern}).Match(parts, false) {
				matches = append(matches, CommitIgnoreMatch{File: file, Pattern: patterns[i]})
				break
			}
		}
	}
	return matches, nil
}

func parseCommitIgnorePatterns(raw []string, field string) ([]gitignore.Pattern, error) {
	parsed := make([]gitignore.Pattern, len(raw))
	for i, pattern := range raw {
		if strings.TrimSpace(pattern) == "" {
			return nil, fmt.Errorf("%s: pattern #%d is empty", field, i+1)
		}
		parsed[i] = gitignore.ParsePattern(pattern, nil)
	}
	return parsed, nil
}
