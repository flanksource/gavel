package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeOverallScore(t *testing.T) {
	tests := []struct {
		name     string
		result   VerifyResult
		expected int
	}{
		{
			name:     "empty result",
			result:   VerifyResult{},
			expected: 0,
		},
		{
			name: "all checks pass, perfect ratings, complete",
			result: VerifyResult{
				Checks: map[string]CheckResult{
					"tests-added": {Pass: true},
					"null-safety": {Pass: true},
				},
				Ratings: map[string]RatingResult{
					"security":      {Score: 100},
					"test_coverage": {Score: 100},
				},
				Completeness: CompletenessResult{Pass: true},
			},
			expected: 100,
		},
		{
			name: "half checks fail, mixed ratings, incomplete",
			result: VerifyResult{
				Checks: map[string]CheckResult{
					"tests-added": {Pass: true},
					"null-safety": {Pass: false},
				},
				Ratings: map[string]RatingResult{
					"security":      {Score: 60},
					"test_coverage": {Score: 80},
				},
				Completeness: CompletenessResult{Pass: false},
			},
			// checks: 50% * 50 = 25, ratings: 35% * 70 = 24.5, completeness: 15% * 0 = 0
			expected: 50,
		},
		{
			name: "all checks fail",
			result: VerifyResult{
				Checks: map[string]CheckResult{
					"tests-added": {Pass: false},
					"null-safety": {Pass: false},
				},
				Ratings: map[string]RatingResult{
					"security": {Score: 50},
				},
				Completeness: CompletenessResult{Pass: true},
			},
			// checks: 0, ratings: 35% * 50 = 17.5, completeness: 15% * 100 = 15
			expected: 33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeOverallScore(tt.result)
			if got != tt.expected {
				t.Errorf("ComputeOverallScore() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestEnabledChecks(t *testing.T) {
	t.Run("no disabled returns all", func(t *testing.T) {
		checks := EnabledChecks(ChecksConfig{})
		if len(checks) != len(AllChecks) {
			t.Errorf("got %d checks, want %d", len(checks), len(AllChecks))
		}
	})

	t.Run("disable by ID", func(t *testing.T) {
		checks := EnabledChecks(ChecksConfig{Disabled: []string{"tests-added", "no-hardcoded-secrets"}})
		if len(checks) != len(AllChecks)-2 {
			t.Errorf("got %d checks, want %d", len(checks), len(AllChecks)-2)
		}
		for _, c := range checks {
			if c.ID == "tests-added" || c.ID == "no-hardcoded-secrets" {
				t.Errorf("disabled check %q should not be present", c.ID)
			}
		}
	})

	t.Run("disable by category", func(t *testing.T) {
		checks := EnabledChecks(ChecksConfig{DisabledCategories: []string{"security"}})
		for _, c := range checks {
			if c.Category == "security" {
				t.Errorf("disabled category check %q should not be present", c.ID)
			}
		}
		if len(checks) >= len(AllChecks) {
			t.Errorf("expected fewer checks after disabling security category")
		}
	})
}

func TestChecksByCategory(t *testing.T) {
	byCategory := ChecksByCategory(AllChecks)
	for _, cat := range AllCategories {
		if len(byCategory[cat]) == 0 {
			t.Errorf("category %q has no checks", cat)
		}
	}
}

func TestMergeVerifyConfig(t *testing.T) {
	base := VerifyConfig{
		Model: "claude",
	}

	t.Run("override model", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{Model: "gemini"})
		if got.Model != "gemini" {
			t.Errorf("Model = %q, want %q", got.Model, "gemini")
		}
	})

	t.Run("merge disabled checks", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{
			Checks: ChecksConfig{Disabled: []string{"tests-added"}},
		})
		if len(got.Checks.Disabled) != 1 || got.Checks.Disabled[0] != "tests-added" {
			t.Errorf("Checks.Disabled = %v, want [tests-added]", got.Checks.Disabled)
		}
	})

	t.Run("empty override changes nothing", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{})
		if got.Model != "claude" {
			t.Errorf("Model = %q, want %q", got.Model, "claude")
		}
	})
}

func TestResolveAdapter(t *testing.T) {
	tests := []struct {
		model         string
		expectedName  string
		expectedModel string
	}{
		{"claude", "claude", "claude"},
		{"gemini", "gemini", "gemini"},
		{"codex", "codex", "codex"},
		{"claude-sonnet-4", "claude", "claude-sonnet-4"},
		{"gemini-2.5-flash", "gemini", "gemini-2.5-flash"},
		{"codex-mini", "codex", "codex-mini"},
		{"unknown", "claude", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			adapter, model := ResolveAdapter(tt.model)
			if adapter.Name() != tt.expectedName {
				t.Errorf("Name() = %q, want %q", adapter.Name(), tt.expectedName)
			}
			if model != tt.expectedModel {
				t.Errorf("model = %q, want %q", model, tt.expectedModel)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfgData := []byte("verify:\n  model: gemini\n  checks:\n    disabled:\n      - tests-added\n")
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Model != "gemini" {
		t.Errorf("Model = %q, want %q", cfg.Model, "gemini")
	}
	if len(cfg.Checks.Disabled) != 1 || cfg.Checks.Disabled[0] != "tests-added" {
		t.Errorf("Checks.Disabled = %v, want [tests-added]", cfg.Checks.Disabled)
	}
}

func TestResolveScope(t *testing.T) {
	t.Run("no args no range", func(t *testing.T) {
		s, err := ResolveScope(nil, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "diff", s.Type)
	})

	t.Run("commit range flag", func(t *testing.T) {
		s, err := ResolveScope(nil, "main..HEAD", "")
		assert.NoError(t, err)
		assert.Equal(t, "range", s.Type)
		assert.Equal(t, "main..HEAD", s.CommitRange)
	})

	t.Run("range flag takes precedence over args", func(t *testing.T) {
		s, err := ResolveScope([]string{"a.go"}, "main..HEAD", "")
		assert.NoError(t, err)
		assert.Equal(t, "range", s.Type)
	})
}

func TestBuildSchema(t *testing.T) {
	checks := EnabledChecks(ChecksConfig{})
	schemaJSON, err := BuildSchema(checks)
	if err != nil {
		t.Fatalf("BuildSchema() error: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatalf("BuildSchema output is not valid JSON: %v", err)
	}

	props := schema["properties"].(map[string]any)
	for _, field := range []string{"checks", "ratings", "completeness"} {
		if _, ok := props[field]; !ok {
			t.Errorf("schema missing property %q", field)
		}
	}

	checksObj := props["checks"].(map[string]any)
	required := checksObj["required"].([]any)
	if len(required) != len(checks) {
		t.Errorf("checks required has %d entries, want %d", len(required), len(checks))
	}
}

func TestBuildSchemaWithDisabled(t *testing.T) {
	checks := EnabledChecks(ChecksConfig{Disabled: []string{"tests-added"}})
	schemaJSON, err := BuildSchema(checks)
	if err != nil {
		t.Fatalf("BuildSchema() error: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	checksObj := schema["properties"].(map[string]any)["checks"].(map[string]any)
	checkProps := checksObj["properties"].(map[string]any)
	if _, ok := checkProps["tests-added"]; ok {
		t.Error("disabled check tests-added should not appear in schema")
	}
}

func TestSchemaFile(t *testing.T) {
	path, err := SchemaFile(ChecksConfig{})
	if err != nil {
		t.Fatalf("SchemaFile() error: %v", err)
	}
	defer os.Remove(path)

	if !strings.HasSuffix(path, ".json") {
		t.Errorf("schema file path should end in .json, got %q", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema file is not valid JSON: %v", err)
	}
}
