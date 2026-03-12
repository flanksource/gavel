package git

import (
	"testing"

	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyConfigFilters_AuthorSkip(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreAuthors: []string{"dependabot*"},
	}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dependabot[bot]", Email: "dependabot@github.com"},
		Subject: "chore(deps): bump something",
	}
	changes := []CommitChange{{File: "go.mod", Adds: 1, Dels: 1}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
	assert.Contains(t, result.Reason, "author")
}

func TestApplyConfigFilters_CommitMessageSkip(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreCommits: []string{"fixup!*"},
	}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "fixup! some previous commit",
	}
	changes := []CommitChange{{File: "main.go", Adds: 5}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
	assert.Contains(t, result.Reason, "commit message")
}

func TestApplyConfigFilters_CommitTypeSkip(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreCommitTypes: []string{"chore", "ci"},
	}

	commit := Commit{
		Hash:       "abc12345",
		Author:     Author{Name: "dev"},
		Subject:    "update CI config",
		CommitType: CommitType("ci"),
	}
	changes := []CommitChange{{File: ".github/workflows/ci.yml", Adds: 10}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
	assert.Contains(t, result.Reason, "commit type")
}

func TestApplyConfigFilters_FileFilter(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreFiles: []string{"*.lock", "go.sum"},
	}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "feat: add feature",
	}
	changes := []CommitChange{
		{File: "main.go", Adds: 50},
		{File: "go.sum", Adds: 100},
		{File: "yarn.lock", Adds: 200},
	}

	result := ApplyConfigFilters(config, commit, changes)
	assert.False(t, result.SkipCommit)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "main.go", result.Changes[0].File)
	assert.Equal(t, 2, result.FilesSkipped)
}

func TestApplyConfigFilters_AllFilesFiltered(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreFiles: []string{"*.lock", "package-lock.json"},
	}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "chore: update lockfile",
	}
	changes := []CommitChange{
		{File: "package-lock.json", Adds: 500},
		{File: "yarn.lock", Adds: 200},
	}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
	assert.Contains(t, result.Reason, "all file changes filtered")
}

func TestApplyConfigFilters_CELRule(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreCommitRules: []CommitRule{
			{CEL: "commit.is_merge"},
		},
	}
	require.NoError(t, config.Compile())

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "Merge branch 'feature' into main",
	}
	changes := []CommitChange{{File: "main.go", Adds: 10}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
	assert.Contains(t, result.Reason, "CEL rule")
}

func TestApplyConfigFilters_CELRuleLineChanges(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreCommitRules: []CommitRule{
			{CEL: "commit.line_changes > 100"},
		},
	}
	require.NoError(t, config.Compile())

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "feat: big change",
	}
	changes := []CommitChange{{File: "main.go", Adds: 80, Dels: 30}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.True(t, result.SkipCommit)
}

func TestApplyConfigFilters_ResourceFilter(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreResources: []ResourceFilter{
			{Kind: "Secret"},
			{Kind: "ConfigMap", Name: "*-generated"},
		},
	}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "feat: update resources",
	}
	changes := []CommitChange{
		{
			File: "deploy/resources.yaml",
			Adds: 20,
			KubernetesChanges: []kubernetes.KubernetesChange{
				{KubernetesRef: kubernetes.KubernetesRef{Kind: "Secret", Name: "my-secret"}},
				{KubernetesRef: kubernetes.KubernetesRef{Kind: "Deployment", Name: "my-app"}},
				{KubernetesRef: kubernetes.KubernetesRef{Kind: "ConfigMap", Name: "app-generated"}},
			},
		},
	}

	result := ApplyConfigFilters(config, commit, changes)
	assert.False(t, result.SkipCommit)
	require.Len(t, result.Changes, 1)
	assert.Len(t, result.Changes[0].KubernetesChanges, 1)
	assert.Equal(t, "Deployment", result.Changes[0].KubernetesChanges[0].Kind)
	assert.Equal(t, 2, result.ResourcesSkipped)
}

func TestApplyConfigFilters_NoFilters(t *testing.T) {
	config := &GitAnalyzeConfig{}

	commit := Commit{
		Hash:    "abc12345",
		Author:  Author{Name: "dev"},
		Subject: "feat: add feature",
	}
	changes := []CommitChange{{File: "main.go", Adds: 10}}

	result := ApplyConfigFilters(config, commit, changes)
	assert.False(t, result.SkipCommit)
	assert.Len(t, result.Changes, 1)
}

func TestApplyConfigFilters_PassesThrough(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreAuthors:     []string{"bot*"},
		IgnoreCommits:     []string{"fixup!*"},
		IgnoreCommitTypes: []string{"chore"},
		IgnoreFiles:       []string{"*.lock"},
	}

	commit := Commit{
		Hash:       "abc12345",
		Author:     Author{Name: "developer"},
		Subject:    "feat: new feature",
		CommitType: CommitType("feat"),
	}
	changes := []CommitChange{
		{File: "main.go", Adds: 10},
		{File: "test.go", Adds: 20},
	}

	result := ApplyConfigFilters(config, commit, changes)
	assert.False(t, result.SkipCommit)
	assert.Len(t, result.Changes, 2)
}

func TestApplyConfigFilters_CommitTypeNegation(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreCommitTypes: []string{"!feat"},
	}

	feat := Commit{Hash: "abc12345", Author: Author{Name: "dev"}, Subject: "feat: add", CommitType: CommitType("feat")}
	chore := Commit{Hash: "def12345", Author: Author{Name: "dev"}, Subject: "chore: cleanup", CommitType: CommitType("chore")}
	changes := []CommitChange{{File: "main.go", Adds: 10}}

	// "!feat" means: skip everything EXCEPT feat
	assert.False(t, ApplyConfigFilters(config, feat, changes).SkipCommit, "feat should NOT be skipped")
	assert.True(t, ApplyConfigFilters(config, chore, changes).SkipCommit, "chore should be skipped")
}

func TestApplyConfigFilters_FileNegation(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreFiles: []string{"*.go", "!*_test.go"},
	}

	commit := Commit{Hash: "abc12345", Author: Author{Name: "dev"}, Subject: "feat: add"}
	changes := []CommitChange{
		{File: "main.go", Adds: 10},
		{File: "main_test.go", Adds: 20},
		{File: "README.md", Adds: 5},
	}

	result := ApplyConfigFilters(config, commit, changes)
	assert.False(t, result.SkipCommit)
	assert.Len(t, result.Changes, 2)
	files := []string{result.Changes[0].File, result.Changes[1].File}
	assert.Contains(t, files, "main_test.go")
	assert.Contains(t, files, "README.md")
	assert.Equal(t, 1, result.FilesSkipped)
}

func TestApplyConfigFilters_Combined(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreAuthors:     []string{"dependabot*"},
		IgnoreFiles:       []string{"*.lock", "go.sum"},
		IgnoreCommitTypes: []string{"chore"},
		IgnoreResources: []ResourceFilter{
			{Kind: "Secret"},
		},
	}

	// Bot commit should be skipped entirely
	botCommit := Commit{Hash: "aaa12345", Author: Author{Name: "dependabot[bot]"}, Subject: "chore(deps): bump"}
	botChanges := []CommitChange{{File: "go.mod", Adds: 1}}
	assert.True(t, ApplyConfigFilters(config, botCommit, botChanges).SkipCommit)

	// Chore commit from human should also be skipped
	choreCommit := Commit{Hash: "bbb12345", Author: Author{Name: "dev"}, CommitType: CommitType("chore"), Subject: "chore: tidy"}
	choreChanges := []CommitChange{{File: "main.go", Adds: 1}}
	assert.True(t, ApplyConfigFilters(config, choreCommit, choreChanges).SkipCommit)

	// Normal commit: lock files removed, resources trimmed, source files kept
	normalCommit := Commit{Hash: "ccc12345", Author: Author{Name: "dev"}, CommitType: CommitType("feat"), Subject: "feat: deploy"}
	normalChanges := []CommitChange{
		{File: "main.go", Adds: 50},
		{File: "go.sum", Adds: 100},
		{File: "yarn.lock", Adds: 200},
		{
			File: "deploy/app.yaml",
			Adds: 20,
			KubernetesChanges: []kubernetes.KubernetesChange{
				{KubernetesRef: kubernetes.KubernetesRef{Kind: "Secret", Name: "db-creds"}},
				{KubernetesRef: kubernetes.KubernetesRef{Kind: "Deployment", Name: "app"}},
			},
		},
	}

	result := ApplyConfigFilters(config, normalCommit, normalChanges)
	assert.False(t, result.SkipCommit)
	assert.Equal(t, 2, result.FilesSkipped)
	assert.Equal(t, 1, result.ResourcesSkipped)
	require.Len(t, result.Changes, 2)
	assert.Equal(t, "main.go", result.Changes[0].File)
	assert.Equal(t, "deploy/app.yaml", result.Changes[1].File)
	assert.Len(t, result.Changes[1].KubernetesChanges, 1)
	assert.Equal(t, "Deployment", result.Changes[1].KubernetesChanges[0].Kind)
}

func TestMatchesFile_NegationPreservesTestFiles(t *testing.T) {
	config := &GitAnalyzeConfig{
		IgnoreFiles: []string{"*.go", "!*_test.go"},
	}
	commit := Commit{Hash: "abc12345", Author: Author{Name: "dev"}, Subject: "test"}

	tests := []struct {
		file     string
		filtered bool
	}{
		{"main.go", true},
		{"main_test.go", false},
		{"handler.go", true},
		{"handler_test.go", false},
		{"README.md", false},
	}

	for _, tc := range tests {
		result := ApplyConfigFilters(config, commit, []CommitChange{{File: tc.file, Adds: 1}})
		if tc.filtered {
			assert.True(t, result.SkipCommit || len(result.Changes) == 0, "expected %s to be filtered", tc.file)
		} else {
			assert.False(t, result.SkipCommit, "expected %s to NOT be filtered", tc.file)
			assert.Len(t, result.Changes, 1, "expected %s to pass through", tc.file)
		}
	}
}
