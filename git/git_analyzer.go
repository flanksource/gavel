package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/flanksource/gavel/models"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/llm"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git/kubernetes"
)

func AnalyzeCommit(ctx *AnalyzerContext, commit Commit, options AnalyzeOptions) (CommitAnalysis, error) {
	out := CommitAnalysis{
		Commit:   commit,
		Original: commit,
	}
	changes, err := ParsePatch(commit.Patch)
	if err != nil {
		return out, fmt.Errorf("failed to parse changes for commit %s: %w", commit.Hash, err)
	}

	// Set arch config from context for filtering
	if ctx.Arch != nil {
		options.arch = *ctx.Arch
	}

	for _, change := range changes {
		conf, err := ctx.GetFileMap(change.File, commit.Hash)
		if err != nil {
			return out, err
		}
		change.Scope = conf.Scopes
		change.Tech = conf.Tech

		// Analyze Kubernetes resources if applicable
		if err := kubernetes.AnalyzeKubernetesChanges(ctx, commit, &change); err != nil {
			return out, err
		}

		if options.Matches(commit, change) {
			out.Changes = append(out.Changes, change)
		}
	}
	if len(out.Changes) == 0 {
		return out, nil
	}

	// Aggregate scope from changes if commit doesn't have one
	if out.Scope == ScopeTypeUnknown && len(out.Changes) > 0 {
		scopeCounts := make(map[ScopeType]int)
		for _, scope := range out.GetScopes() {
			if scope != ScopeTypeUnknown {
				scopeCounts[scope]++
			}
		}
		if len(scopeCounts) > 0 {
			out.Scope = findMostCommonScope(scopeCounts)
		}
	}

	// Aggregate tech from all changes
	techSet := make(map[ScopeTechnology]struct{}, len(out.Changes))
	for _, change := range out.Changes {
		for _, tech := range change.Tech {
			techSet[tech] = struct{}{}
		}
	}
	if len(techSet) > 0 {
		out.Tech = make([]ScopeTechnology, 0, len(techSet))
		for tech := range techSet {
			out.Tech = append(out.Tech, tech)
		}
	}

	out.QualityScore = GetQualityScore(out)

	// Pre-compute metrics for performance
	for _, change := range out.Changes {
		out.TotalLineChanges += change.Adds + change.Dels
		out.TotalResourceCount += len(change.KubernetesChanges)
	}

	if options.agent != nil {
		out, err = AnalyzeWithAI(context.Background(), out, options.agent, options)
	}

	return out, err
}

func AnalyzeCommitHistory(ctx *AnalyzerContext, commits []Commit, options AnalyzeOptions) (CommitAnalyses, error) {

	if options.AITimeout.Milliseconds() == 0 {
		options.AITimeout = 2 * time.Minute
	}
	start := time.Now()
	batch := task.Batch[CommitAnalysis]{
		Name:        "Analyze Commit History",
		ItemTimeout: options.AITimeout,
	}

	if options.AI {

		agent, err := llm.NewLLMAgent(ai.DefaultConfig())
		defer func() {
			logger.Infof("AI Costs: %s", agent.GetCosts().Pretty().ANSI())
		}()
		// agent, err := ai.GetDefaultAgent()
		if err != nil {
			return nil, fmt.Errorf("failed to get AI agent: %w", err)
		}
		options.agent = agent
	}

	for _, commit := range commits {
		commit := commit
		batch.Items = append(batch.Items, func(logger logger.Logger) (CommitAnalysis, error) {
			logger.Infof("analyzing %s with timeout %v", commit.PrettyShort().ANSI(), options.AITimeout)
			analysis, err := AnalyzeCommit(ctx, commit, options)
			if err != nil {
				return CommitAnalysis{}, fmt.Errorf("failed to analyze commit %s: %w", commit.Hash, err)
			}
			return analysis, nil
		})
	}
	results := make(CommitAnalyses, 0, len(commits))
	var err error
	for item := range batch.Run() {
		results = append(results, item.Value)
		if item.Error != nil {
			err = fmt.Errorf("failed to analyze some commits: %w", item.Error)
		}
	}
	logger.Infof("analyzed %d commits in %v", len(results), time.Since(start))
	return results, err
}

func ReadFileAtCommit(filter HistoryOptions, commit, path string) (string, error) {
	if filter.Path == "" {
		wd, _ := os.Getwd()
		filter.Path = wd
	}

	return "", nil

}

// findMostCommonScope returns the scope that appears most frequently in the given counts
func findMostCommonScope(scopeCounts map[ScopeType]int) ScopeType {
	var mostCommon ScopeType
	maxCount := 0
	for scope, count := range scopeCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = scope
		}
	}
	return mostCommon
}

// LoadCommitAnalysesFromJSON loads commit analyses from multiple JSON files and merges them
func LoadCommitAnalysesFromJSON(filePaths []string) (CommitAnalyses, error) {
	var allAnalyses CommitAnalyses

	for _, filePath := range filePaths {
		logger.Infof("Loading commit analyses from %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		var analyses CommitAnalyses
		if err := json.Unmarshal(data, &analyses); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
		}

		// Validate that repository field is present
		for i, analysis := range analyses {
			if analysis.Repository == "" {
				return nil, fmt.Errorf("commit at index %d in file %s is missing repository field", i, filePath)
			}
		}

		allAnalyses = append(allAnalyses, analyses...)
		logger.Infof("Loaded %d commits from %s", len(analyses), filePath)
	}

	logger.Infof("Total commits loaded: %d", len(allAnalyses))
	return allAnalyses, nil
}

// getRepositoryName attempts to determine the repository name from git remote or falls back to directory name
func getRepositoryName(path string) string {
	// Try to get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err == nil {
		remoteURL := strings.TrimSpace(string(output))
		// Parse repository name from URL
		// Examples:
		// - https://github.com/org/repo.git -> repo
		// - git@github.com:org/repo.git -> repo
		parts := strings.Split(remoteURL, "/")
		if len(parts) > 0 {
			repoName := parts[len(parts)-1]
			repoName = strings.TrimSuffix(repoName, ".git")
			if repoName != "" {
				return repoName
			}
		}
	}

	// Fallback to directory name
	absPath, err := filepath.Abs(path)
	if err == nil {
		return filepath.Base(absPath)
	}

	return filepath.Base(path)
}

// ApplyFilters applies HistoryOptions filters to a slice of commits
func ApplyFilters(commits CommitAnalyses, filter HistoryOptions) CommitAnalyses {
	var filtered CommitAnalyses
	for _, commit := range commits {
		if filter.Matches(commit.Commit) {
			filtered = append(filtered, commit)
		}
	}
	return filtered
}
