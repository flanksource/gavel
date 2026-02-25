package types

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTODOVerifyConfig_Serialization(t *testing.T) {
	input := `title: test
priority: high
status: pending
language: go
verify:
  categories:
    - testing
    - code-quality
  score_threshold: 85
`
	var fm TODOFrontmatter
	if err := yaml.Unmarshal([]byte(input), &fm); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if fm.Verify == nil {
		t.Fatal("Expected Verify to be parsed, got nil")
	}
	if len(fm.Verify.Categories) != 2 {
		t.Fatalf("Expected 2 categories, got %d", len(fm.Verify.Categories))
	}
	if fm.Verify.Categories[0] != "testing" || fm.Verify.Categories[1] != "code-quality" {
		t.Errorf("Unexpected categories: %v", fm.Verify.Categories)
	}
	if fm.Verify.ScoreThreshold != 85 {
		t.Errorf("Expected score_threshold 85, got %d", fm.Verify.ScoreThreshold)
	}

	out, err := yaml.Marshal(&fm)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var roundTripped TODOFrontmatter
	if err := yaml.Unmarshal(out, &roundTripped); err != nil {
		t.Fatalf("Failed to round-trip unmarshal: %v", err)
	}
	if roundTripped.Verify == nil {
		t.Fatal("Expected Verify after round-trip, got nil")
	}
	if roundTripped.Verify.ScoreThreshold != 85 {
		t.Errorf("Round-trip score_threshold: expected 85, got %d", roundTripped.Verify.ScoreThreshold)
	}
}

func TestTODOVerifyConfig_OmittedWhenNil(t *testing.T) {
	fm := TODOFrontmatter{
		Title:    "test",
		Priority: PriorityHigh,
		Status:   StatusPending,
	}

	out, err := yaml.Marshal(&fm)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var roundTripped TODOFrontmatter
	if err := yaml.Unmarshal(out, &roundTripped); err != nil {
		t.Fatalf("Failed to round-trip: %v", err)
	}
	if roundTripped.Verify != nil {
		t.Errorf("Expected Verify to be nil after round-trip, got %+v", roundTripped.Verify)
	}
}

func TestStringOrSlice_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StringOrSlice
	}{
		{"single string", `path: pkg/auth/login.go`, StringOrSlice{"pkg/auth/login.go"}},
		{"list of strings", "path:\n  - pkg/auth/login.go\n  - pkg/auth/session.go", StringOrSlice{"pkg/auth/login.go", "pkg/auth/session.go"}},
		{"glob pattern", `path: "pkg/auth/*.go"`, StringOrSlice{"pkg/auth/*.go"}},
		{"empty omitted", `priority: high`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := "priority: high\nstatus: pending\n" + tc.input
			var fm TODOFrontmatter
			require.NoError(t, yaml.Unmarshal([]byte(input), &fm))
			assert.Equal(t, tc.expected, fm.Path)
		})
	}
}

func TestStringOrSlice_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		path StringOrSlice
	}{
		{"single", StringOrSlice{"pkg/auth/login.go"}},
		{"multiple", StringOrSlice{"pkg/auth/login.go", "pkg/auth/session.go"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fm := TODOFrontmatter{Priority: PriorityHigh, Status: StatusPending, Path: tc.path}
			out, err := yaml.Marshal(&fm)
			require.NoError(t, err)

			var roundTripped TODOFrontmatter
			require.NoError(t, yaml.Unmarshal(out, &roundTripped))
			assert.Equal(t, tc.path, roundTripped.Path)
		})
	}
}

func TestBranch_Serialization(t *testing.T) {
	input := `title: test
priority: high
status: pending
branch: pr/fix-terminal
`
	var fm TODOFrontmatter
	require.NoError(t, yaml.Unmarshal([]byte(input), &fm))
	assert.Equal(t, "pr/fix-terminal", fm.Branch)

	out, err := yaml.Marshal(&fm)
	require.NoError(t, err)

	var roundTripped TODOFrontmatter
	require.NoError(t, yaml.Unmarshal(out, &roundTripped))
	assert.Equal(t, "pr/fix-terminal", roundTripped.Branch)
}

func TestBranch_OmittedWhenEmpty(t *testing.T) {
	fm := TODOFrontmatter{Title: "test", Priority: PriorityHigh, Status: StatusPending}
	out, err := yaml.Marshal(&fm)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "branch")
}

func TestCleanMetadata_RemovesBranchKey(t *testing.T) {
	fm := TODOFrontmatter{Branch: "main"}
	fm.Metadata = map[string]any{"branch": "main", "custom": "kept"}
	fm.CleanMetadata()

	_, exists := fm.Metadata["branch"]
	assert.False(t, exists, "Expected 'branch' to be removed from Metadata")
	assert.Equal(t, "kept", fm.Metadata["custom"])
}

func TestCleanMetadata_RemovesVerifyKey(t *testing.T) {
	fm := TODOFrontmatter{
		Verify: &TODOVerifyConfig{Categories: []string{"testing"}, ScoreThreshold: 80},
	}
	fm.Metadata = map[string]any{
		"verify":   map[string]any{"categories": []string{"testing"}},
		"title":    "test",
		"priority": "high",
		"custom":   "kept",
	}

	fm.CleanMetadata()

	if _, exists := fm.Metadata["verify"]; exists {
		t.Error("Expected 'verify' to be removed from Metadata")
	}
	if _, exists := fm.Metadata["title"]; exists {
		t.Error("Expected 'title' to be removed from Metadata")
	}
	if v, exists := fm.Metadata["custom"]; !exists || v != "kept" {
		t.Error("Expected 'custom' key to remain in Metadata")
	}
}

func TestPR_Serialization(t *testing.T) {
	input := `title: test
priority: high
status: pending
prompt: "Fix the null pointer"
pr:
  number: 42
  url: https://github.com/org/repo/pull/42
  head: feat/review
  base: main
  comment_id: 100
  comment_author: reviewer
  comment_url: https://github.com/org/repo/pull/42#discussion_r100
`
	var fm TODOFrontmatter
	require.NoError(t, yaml.Unmarshal([]byte(input), &fm))

	assert.Equal(t, "Fix the null pointer", fm.Prompt)
	require.NotNil(t, fm.PR)
	assert.Equal(t, 42, fm.PR.Number)
	assert.Equal(t, "https://github.com/org/repo/pull/42", fm.PR.URL)
	assert.Equal(t, "feat/review", fm.PR.Head)
	assert.Equal(t, "main", fm.PR.Base)
	assert.Equal(t, int64(100), fm.PR.CommentID)
	assert.Equal(t, "reviewer", fm.PR.CommentAuthor)
	assert.Equal(t, "https://github.com/org/repo/pull/42#discussion_r100", fm.PR.CommentURL)

	out, err := yaml.Marshal(&fm)
	require.NoError(t, err)

	var roundTripped TODOFrontmatter
	require.NoError(t, yaml.Unmarshal(out, &roundTripped))
	assert.Equal(t, fm.PR, roundTripped.PR)
	assert.Equal(t, fm.Prompt, roundTripped.Prompt)
}

func TestPR_OmittedWhenNil(t *testing.T) {
	fm := TODOFrontmatter{Title: "test", Priority: PriorityHigh, Status: StatusPending}
	out, err := yaml.Marshal(&fm)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "pr:")
	assert.NotContains(t, string(out), "prompt:")
}

func TestCleanMetadata_RemovesPRAndPromptKeys(t *testing.T) {
	fm := TODOFrontmatter{
		Prompt: "fix it",
		PR:     &PR{Number: 42},
	}
	fm.Metadata = map[string]any{"pr": map[string]any{}, "prompt": "fix it", "custom": "kept"}
	fm.CleanMetadata()

	_, prExists := fm.Metadata["pr"]
	_, promptExists := fm.Metadata["prompt"]
	assert.False(t, prExists)
	assert.False(t, promptExists)
	assert.Equal(t, "kept", fm.Metadata["custom"])
}
