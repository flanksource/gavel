package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeOverallScore(t *testing.T) {
	tests := []struct {
		name     string
		sections []SectionResult
		weights  map[string]float64
		expected int
	}{
		{
			name:     "empty sections",
			sections: nil,
			weights:  nil,
			expected: 0,
		},
		{
			name: "equal weights",
			sections: []SectionResult{
				{Name: "security", Score: 80},
				{Name: "performance", Score: 60},
			},
			weights:  map[string]float64{"security": 1.0, "performance": 1.0},
			expected: 70,
		},
		{
			name: "weighted scores",
			sections: []SectionResult{
				{Name: "security", Score: 100},
				{Name: "performance", Score: 50},
			},
			weights:  map[string]float64{"security": 2.0, "performance": 1.0},
			expected: 83,
		},
		{
			name: "missing weight defaults to 1.0",
			sections: []SectionResult{
				{Name: "security", Score: 90},
				{Name: "unknown", Score: 70},
			},
			weights:  map[string]float64{"security": 1.0},
			expected: 80,
		},
		{
			name: "nil weights defaults all to 1.0",
			sections: []SectionResult{
				{Name: "a", Score: 60},
				{Name: "b", Score: 80},
				{Name: "c", Score: 100},
			},
			weights:  nil,
			expected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeOverallScore(tt.sections, tt.weights)
			if got != tt.expected {
				t.Errorf("ComputeOverallScore() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMergeVerifyConfig(t *testing.T) {
	base := VerifyConfig{
		Model:    "claude",
		Sections: []string{"security", "performance"},
		Weights:  map[string]float64{"security": 2.0, "performance": 1.0},
	}

	t.Run("override model", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{Model: "gemini"})
		if got.Model != "gemini" {
			t.Errorf("Model = %q, want %q", got.Model, "gemini")
		}
		if len(got.Sections) != 2 {
			t.Errorf("Sections should be preserved, got %d", len(got.Sections))
		}
	})

	t.Run("override sections", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{Sections: []string{"testing"}})
		if len(got.Sections) != 1 || got.Sections[0] != "testing" {
			t.Errorf("Sections = %v, want [testing]", got.Sections)
		}
	})

	t.Run("merge weights", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{Weights: map[string]float64{"testing": 1.5}})
		if got.Weights["testing"] != 1.5 {
			t.Errorf("Weights[testing] = %f, want 1.5", got.Weights["testing"])
		}
		if got.Weights["security"] != 2.0 {
			t.Errorf("Weights[security] = %f, want 2.0 (should be preserved)", got.Weights["security"])
		}
	})

	t.Run("empty override changes nothing", func(t *testing.T) {
		got := MergeVerifyConfig(base, VerifyConfig{})
		if got.Model != "claude" {
			t.Errorf("Model = %q, want %q", got.Model, "claude")
		}
	})
}

func TestResolveCLI(t *testing.T) {
	tests := []struct {
		model          string
		expectedBinary string
		expectedModel  string
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
			tool, model := ResolveCLI(tt.model)
			if tool.Binary != tt.expectedBinary {
				t.Errorf("Binary = %q, want %q", tool.Binary, tt.expectedBinary)
			}
			if model != tt.expectedModel {
				t.Errorf("model = %q, want %q", model, tt.expectedModel)
			}
		})
	}
}

func TestParseVerifyResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantScore int
		wantCount int
		wantErr   bool
	}{
		{
			name: "plain yaml",
			input: `sections:
  - name: security
    score: 90
  - name: performance
    score: 70`,
			wantCount: 2,
		},
		{
			name:      "yaml with fences",
			input:     "```yaml\nsections:\n  - name: security\n    score: 85\n```",
			wantCount: 1,
		},
		{
			name:      "json wrapper with result field",
			input:     `{"result": "sections:\n  - name: testing\n    score: 75"}`,
			wantCount: 1,
		},
		{
			name:    "invalid input",
			input:   "not yaml at all {{{",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVerifyResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Sections) != tt.wantCount {
				t.Errorf("got %d sections, want %d", len(result.Sections), tt.wantCount)
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

	cfgData := []byte("verify:\n  model: gemini\n  weights:\n    security: 3.0\n")
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
	if cfg.Weights["security"] != 3.0 {
		t.Errorf("Weights[security] = %f, want 3.0", cfg.Weights["security"])
	}
	if len(cfg.Sections) != 6 {
		t.Errorf("Sections should have defaults, got %d", len(cfg.Sections))
	}
}

func TestResolveScope(t *testing.T) {
	t.Run("no args no range", func(t *testing.T) {
		s := ResolveScope(nil, "")
		if s.Type != "diff" {
			t.Errorf("Type = %q, want diff", s.Type)
		}
	})

	t.Run("commit range", func(t *testing.T) {
		s := ResolveScope(nil, "main..HEAD")
		if s.Type != "range" || s.CommitRange != "main..HEAD" {
			t.Errorf("got %+v, want range with main..HEAD", s)
		}
	})

	t.Run("file args", func(t *testing.T) {
		s := ResolveScope([]string{"a.go", "b.go"}, "")
		if s.Type != "files" || len(s.Files) != 2 {
			t.Errorf("got %+v, want files with 2 entries", s)
		}
	})

	t.Run("range takes precedence over files", func(t *testing.T) {
		s := ResolveScope([]string{"a.go"}, "main..HEAD")
		if s.Type != "range" {
			t.Errorf("Type = %q, want range (should take precedence)", s.Type)
		}
	})
}
