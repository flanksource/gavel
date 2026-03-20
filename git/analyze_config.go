package git

import (
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/flanksource/repomap"
	"github.com/google/cel-go/common/types"
)

// FilterResult holds the result of applying exclude filters to a commit
type FilterResult struct {
	SkipCommit       bool
	Reason           string
	Changes          []models.CommitChange
	FilesSkipped     int
	ResourcesSkipped int
}

// ApplyConfigFilters applies all config-based exclude filters to a commit and its changes
func ApplyConfigFilters(config *repomap.CompiledExcludeConfig, commit models.Commit, changes []models.CommitChange) FilterResult {
	// 1. Author filter
	if matched, reason := repomap.MatchesAuthor(toRepomapAuthor(commit.Author), config.Authors); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}
	if matched, reason := repomap.MatchesAuthor(toRepomapAuthor(commit.Committer), config.Authors); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 2. Commit message filter
	if matched, reason := repomap.MatchesCommitMessage(commit.Subject, config.Commits); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 3. Commit type filter
	if matched, reason := repomap.MatchesCommitType(string(commit.CommitType), config.CommitTypes); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 4. CEL commit rules
	if matched, reason := evalCommitCELRules(config, commit, changes); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 5. File path filter + resource filter (per-change)
	var filtered []models.CommitChange
	filesSkipped := 0
	resourcesSkipped := 0

	for _, change := range changes {
		if matched, reason := repomap.MatchesFile(change.File, config.Files); matched {
			filesSkipped++
			if logger.V(3).Enabled() {
				logger.Tracef("Skipping file change: %s (%s)", change.File, reason)
			}
			continue
		}

		// 6. Resource-level filter
		if len(config.Resources) > 0 && len(change.KubernetesChanges) > 0 {
			var keptResources []kubernetes.KubernetesChange
			for _, kc := range change.KubernetesChanges {
				if matchesResource(kc, config.Resources) {
					resourcesSkipped++
					continue
				}
				keptResources = append(keptResources, kc)
			}
			change.KubernetesChanges = keptResources
		}

		filtered = append(filtered, change)
	}

	if len(filtered) == 0 && len(changes) > 0 {
		return FilterResult{
			SkipCommit:   true,
			Reason:       "all file changes filtered",
			FilesSkipped: filesSkipped,
		}
	}

	return FilterResult{
		Changes:          filtered,
		FilesSkipped:     filesSkipped,
		ResourcesSkipped: resourcesSkipped,
	}
}

func matchesResource(kc kubernetes.KubernetesChange, filters []repomap.ResourceFilter) bool {
	for _, f := range filters {
		if f.When != "" {
			// FIXME: CEL resource evaluation not yet implemented
			continue
		}
		if repomap.MatchesResourceFields(kc.Kind, kc.Name, kc.Namespace, f) {
			return true
		}
	}
	return false
}

func evalCommitCELRules(config *repomap.CompiledExcludeConfig, commit models.Commit, changes []models.CommitChange) (bool, string) {
	programs := config.CompiledRules()
	if len(programs) == 0 {
		return false, ""
	}

	ctx := buildFilterCommitContext(commit, changes)

	for i, prog := range programs {
		result, _, err := prog.Eval(ctx)
		if err != nil {
			continue
		}
		if boolVal, ok := result.(types.Bool); ok && boolVal == types.True {
			return true, fmt.Sprintf("CEL rule '%s'", config.Rules[i].When)
		}
	}
	return false, ""
}

func buildFilterCommitContext(commit models.Commit, changes []models.CommitChange) map[string]any {
	files := make([]string, 0, len(changes))
	totalAdds, totalDels := 0, 0
	for _, c := range changes {
		files = append(files, c.File)
		totalAdds += c.Adds
		totalDels += c.Dels
	}

	isMerge := strings.HasPrefix(commit.Subject, "Merge ")

	return map[string]any{
		"commit": map[string]any{
			"hash":          commit.Hash,
			"author":        commit.Author.Name,
			"author_email":  commit.Author.Email,
			"subject":       commit.Subject,
			"message":       commit.Subject,
			"body":          commit.Body,
			"type":          string(commit.CommitType),
			"scope":         string(commit.Scope),
			"is_merge":      isMerge,
			"files_changed": len(changes),
			"line_changes":  totalAdds + totalDels,
			"additions":     totalAdds,
			"deletions":     totalDels,
			"files":         files,
			"tags":          commit.Tags,
			"is_tagged":     len(commit.Tags) > 0,
		},
	}
}

func toRepomapAuthor(a models.Author) repomap.Author {
	return repomap.Author{
		Name:  a.Name,
		Email: a.Email,
		Date:  a.Date,
	}
}
