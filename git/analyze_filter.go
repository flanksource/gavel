package git

import (
	"fmt"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/commons/logger"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
)

type FilterResult struct {
	SkipCommit bool
	Reason     string
	// Filtered changes (files that weren't ignored)
	Changes []CommitChange
	// Number of changes removed
	FilesSkipped int
	// Number of resources removed
	ResourcesSkipped int
}

// ApplyConfigFilters applies all config-based filters to a commit and its changes
func ApplyConfigFilters(config *GitAnalyzeConfig, commit Commit, changes []CommitChange) FilterResult {
	// 1. Author filter
	if matched, reason := matchesAuthor(commit, config.IgnoreAuthors); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 2. Commit message filter
	if matched, reason := matchesCommitMessage(commit.Subject, config.IgnoreCommits); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 3. Commit type filter
	if matched, reason := matchesCommitType(commit.CommitType, config.IgnoreCommitTypes); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 4. CEL commit rules
	if matched, reason := evalCommitCELRules(config.IgnoreCommitRules, config.compiledCommitRules, commit, changes); matched {
		return FilterResult{SkipCommit: true, Reason: reason}
	}

	// 5. File path filter + resource filter (per-change)
	var filtered []CommitChange
	filesSkipped := 0
	resourcesSkipped := 0

	for _, change := range changes {
		if matched, reason := matchesFile(change.File, config.IgnoreFiles); matched {
			filesSkipped++
			if logger.V(3).Enabled() {
				logger.Tracef("Skipping file change: %s (%s)", change.File, reason)
			}
			continue
		}

		// 6. Resource-level filter
		if len(config.IgnoreResources) > 0 && len(change.KubernetesChanges) > 0 {
			var keptResources []kubernetes.KubernetesChange
			for _, kc := range change.KubernetesChanges {
				if matched, _ := matchesResource(kc, config.IgnoreResources); matched {
					resourcesSkipped++
					continue
				}
				keptResources = append(keptResources, kc)
			}
			change.KubernetesChanges = keptResources
		}

		filtered = append(filtered, change)
	}

	// If all changes were filtered, skip the commit
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

func matchesResource(kc kubernetes.KubernetesChange, filters []ResourceFilter) (bool, string) {
	for _, f := range filters {
		if f.CEL != "" {
			// FIXME: CEL resource evaluation not yet implemented
			continue
		}
		if matchesResourceFields(kc, f) {
			return true, fmt.Sprintf("resource %s/%s matches filter", kc.Kind, kc.Name)
		}
	}
	return false, ""
}

func matchesResourceFields(kc kubernetes.KubernetesChange, f ResourceFilter) bool {
	if f.Kind != "" {
		if matched, _ := collections.MatchItem(kc.Kind, f.Kind); !matched {
			return false
		}
	}
	if f.Name != "" {
		if matched, _ := collections.MatchItem(kc.Name, f.Name); !matched {
			return false
		}
	}
	if f.Namespace != "" {
		if matched, _ := collections.MatchItem(kc.Namespace, f.Namespace); !matched {
			return false
		}
	}
	return true
}
