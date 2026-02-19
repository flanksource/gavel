package types

import (
	"testing"

	"github.com/ghodss/yaml"
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
