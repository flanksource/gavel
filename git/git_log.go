package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
)

func GetCommitHistory(filter HistoryOptions) (Commits, error) {
	if filter.Path == "" {
		wd, _ := os.Getwd()
		filter.Path = wd
		logger.Debugf("GetCommitHistory: path not set, using current directory: %s", filter.Path)
	}

	// Parse Args into typed fields (CommitShas, CommitRanges, FilePaths)
	if err := filter.ParseArgs(); err != nil {
		return nil, err
	}

	clicky.Infof("git log %s", filter.Pretty().ANSI())

	// Detect repository name once for all commits
	repoName := getRepositoryName(filter.Path)
	start := time.Now()

	var allCommits []Commit

	// Mode 1: Specific commits provided
	if len(filter.CommitShas) > 0 {
		for _, sha := range filter.CommitShas {
			commit, err := getCommitBySHA(filter.Path, sha, filter.FilePaths)
			if err != nil {
				return nil, fmt.Errorf("invalid commit %s: %w", sha, err)
			}
			allCommits = append(allCommits, commit)
		}
	}

	// Mode 2: Commit ranges provided
	for _, rng := range filter.CommitRanges {
		rangeCommits, err := getCommitsByRange(filter.Path, rng, filter.FilePaths)
		if err != nil {
			return nil, fmt.Errorf("invalid range %s: %w", rng, err)
		}
		allCommits = append(allCommits, rangeCommits...)
	}

	// Mode 3: Use git log with filters (existing behavior)
	if len(filter.CommitShas) == 0 && len(filter.CommitRanges) == 0 {
		// Build git log command
		args := []string{
			"log",
			"--all",
			"--date=iso-strict",
			"--pretty=format:%x1e%H%x1f%an%x1f%ae%x1f%aI%x1f%cn%x1f%ce%x1f%cI%x1f%s%x1f%b%x1f%(trailers:unfold,separator=%x1d,keyonly)%x1f%(trailers:unfold,separator=%x1d,valueonly)%x00",
		}

		// Apply date filters
		if !filter.Since.IsZero() {
			args = append(args, fmt.Sprintf("--since=%s", filter.Since.Format(time.RFC3339)))
		}
		if !filter.Until.IsZero() {
			args = append(args, fmt.Sprintf("--until=%s", filter.Until.Format(time.RFC3339)))
		}

		// Apply author filters
		for _, author := range filter.Author {
			args = append(args, fmt.Sprintf("--author=%s", author))
		}

		// Apply message filter
		if filter.Message != "" {
			args = append(args, fmt.Sprintf("--grep=%s", filter.Message))
		}

		// Only include patch data when ShowPatch is true
		if filter.ShowPatch {
			args = append(args, "-p")
		}

		// Add path filters if provided
		if len(filter.FilePaths) > 0 {
			args = append(args, "--")
			args = append(args, filter.FilePaths...)
		}

		logger.Tracef("git %s", strings.Join(args, " "))

		// Execute git command
		cmd := exec.Command("git", args...)
		cmd.Dir = filter.Path
		output, err := cmd.CombinedOutput()
		if err != nil {
			clicky.Errorf("git command failed: %v\nOutput: %s", err, string(output))
			return nil, fmt.Errorf("failed to execute git log: %w\nOutput: %s", err, string(output))
		}

		// Parse git output
		commits, err := ParseGitLogOutput(output)
		if err != nil {
			logger.Errorf("GetCommitHistory: failed to parse git output: %v", err)
			return nil, fmt.Errorf("failed to parse git log output: %w", err)
		}
		allCommits = append(allCommits, commits...)
	}

	if logger.V(3).Enabled() {
		logger.Tracef(clicky.MustFormat(allCommits, clicky.FormatOptions{Pretty: true}))
	}

	// Apply additional filters that aren't handled by git CLI
	var commits []Commit
	for _, commit := range allCommits {
		if filter.Matches(commit) {
			commit.Repository = repoName
			commits = append(commits, commit)
		}
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits match the specified criteria")
	}

	clicky.Infof("git log retrieved %d commits (%d after filtering) in %v for repository %s", len(allCommits), len(commits), time.Since(start), repoName)

	return commits, nil
}

func ParseGitLogOutput(output []byte) ([]Commit, error) {
	if len(output) == 0 {
		return nil, nil
	}

	const (
		RS  = "\x1e" // Record separator (between commits)
		US  = "\x1f" // Unit separator (between fields)
		GS  = "\x1d" // Group separator (between trailer keys/values)
		NUL = "\x00" // Null (between header and patch)
	)

	commitRecords := strings.Split(string(output), RS)
	var commits []Commit

	for _, record := range commitRecords {
		if strings.TrimSpace(record) == "" {
			continue
		}

		// Split header from patch
		parts := strings.SplitN(record, NUL, 2)
		headerStr := parts[0]
		patchStr := ""
		if len(parts) > 1 {
			patchStr = parts[1]
		}

		// Parse header fields
		fields := strings.Split(headerStr, US)
		if len(fields) < 11 {
			logger.Tracef("skipping malformed commit record with %d fields (expected 11)", len(fields))
			continue
		}

		// Parse dates
		authorDate, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			logger.Tracef("failed to parse author date '%s': %v", fields[3], err)
			authorDate = time.Time{}
		}

		committerDate, err := time.Parse(time.RFC3339, fields[6])
		if err != nil {
			logger.Tracef("failed to parse committer date '%s': %v", fields[6], err)
			committerDate = time.Time{}
		}

		// Parse trailers
		trailerKeys := strings.Split(fields[9], GS)
		trailerValues := strings.Split(fields[10], GS)
		trailers := make(map[string]string)
		for i := 0; i < len(trailerKeys) && i < len(trailerValues); i++ {
			key := strings.TrimSpace(trailerKeys[i])
			value := strings.TrimSpace(trailerValues[i])
			if key == "" {
				continue
			}
			// Concatenate duplicate keys with comma
			if existing, ok := trailers[key]; ok {
				trailers[key] = existing + ", " + value
			} else {
				trailers[key] = value
			}
		}

		// Build commit message for parsing
		message := fields[7] // subject
		if fields[8] != "" { // body
			message += "\n" + fields[8]
		}

		// Use NewCommit to parse conventional commit format, tags, references
		commit := NewCommit(message)
		commit.Hash = fields[0]
		commit.Author = Author{
			Name:  fields[1],
			Email: fields[2],
			Date:  authorDate,
		}
		commit.Committer = Author{
			Name:  fields[4],
			Email: fields[5],
			Date:  committerDate,
		}
		commit.Patch = patchStr

		// Merge trailers from git with those parsed from body
		for k, v := range trailers {
			if existing, ok := commit.Trailers[k]; ok {
				commit.Trailers[k] = existing + ", " + v
			} else {
				commit.Trailers[k] = v
			}
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// getCommitBySHA fetches a single commit by SHA using git show
func getCommitBySHA(repoPath, sha string, pathFilters []string) (Commit, error) {
	args := []string{
		"show",
		"--date=iso-strict",
		"--pretty=format:%x1e%H%x1f%an%x1f%ae%x1f%aI%x1f%cn%x1f%ce%x1f%cI%x1f%s%x1f%b%x1f%(trailers:unfold,separator=%x1d,keyonly)%x1f%(trailers:unfold,separator=%x1d,valueonly)%x00",
		"-p",
		sha,
	}

	// Add path filters if provided
	if len(pathFilters) > 0 {
		args = append(args, "--")
		args = append(args, pathFilters...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Commit{}, fmt.Errorf("failed to get commit %s: %w", sha, err)
	}

	commits, err := ParseGitLogOutput(output)
	if err != nil || len(commits) == 0 {
		return Commit{}, fmt.Errorf("failed to parse commit %s: %w", sha, err)
	}

	return commits[0], nil
}

// getCommitsByRange fetches commits in a range using git log
func getCommitsByRange(repoPath, commitRange string, pathFilters []string) ([]Commit, error) {
	args := []string{
		"log",
		"--date=iso-strict",
		"--pretty=format:%x1e%H%x1f%an%x1f%ae%x1f%aI%x1f%cn%x1f%ce%x1f%cI%x1f%s%x1f%b%x1f%(trailers:unfold,separator=%x1d,keyonly)%x1f%(trailers:unfold,separator=%x1d,valueonly)%x00",
		"-p",
		commitRange,
	}

	// Add path filters if provided
	if len(pathFilters) > 0 {
		args = append(args, "--")
		args = append(args, pathFilters...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits for range %s: %w", commitRange, err)
	}

	commits, err := ParseGitLogOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commits for range %s: %w", commitRange, err)
	}

	return commits, nil
}
