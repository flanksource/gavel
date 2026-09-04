package verify

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/captain/pkg/api"
)

func TestPromptSpec_Unmarshal(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		check func(t *testing.T, s PromptSpec)
	}{
		{
			name: "bare string is prompt body",
			in:   `"Write a commit message for {{diff}}"`,
			check: func(t *testing.T, s PromptSpec) {
				if s.Spec.Prompt.User != "Write a commit message for {{diff}}" {
					t.Errorf("User = %q", s.Spec.Prompt.User)
				}
			},
		},
		{
			name: "object form sets model and prompt",
			in:   `{"model":"claude-sonnet-4-5","prompt":{"user":"body"},"budget":{"cost":2}}`,
			check: func(t *testing.T, s PromptSpec) {
				if s.Spec.Model.Name != "claude-sonnet-4-5" {
					t.Errorf("Model = %q", s.Spec.Model.Name)
				}
				if s.Spec.Prompt.User != "body" || s.Spec.Budget.Cost != 2 {
					t.Errorf("prompt/budget = %+v / %+v", s.Spec.Prompt, s.Spec.Budget)
				}
			},
		},
		{
			name: "object form with prompt string shorthand",
			in:   `{"model":"m","prompt":"just the body"}`,
			check: func(t *testing.T, s PromptSpec) {
				if s.Spec.Model.Name != "m" || s.Spec.Prompt.User != "just the body" {
					t.Errorf("got model=%q user=%q", s.Spec.Model.Name, s.Spec.Prompt.User)
				}
			},
		},
		{
			name: "string containing inline JSON spec is parsed",
			in:   `"{\"model\":\"inline-model\",\"prompt\":{\"user\":\"b\"}}"`,
			check: func(t *testing.T, s PromptSpec) {
				if s.Spec.Model.Name != "inline-model" || s.Spec.Prompt.User != "b" {
					t.Errorf("got model=%q user=%q", s.Spec.Model.Name, s.Spec.Prompt.User)
				}
			},
		},
		{
			name: "file reference",
			in:   `{"file":".gavel/prompts/x.prompt"}`,
			check: func(t *testing.T, s PromptSpec) {
				if s.File != ".gavel/prompts/x.prompt" {
					t.Errorf("File = %q", s.File)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s PromptSpec
			if err := json.Unmarshal([]byte(tc.in), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			tc.check(t, s)
		})
	}
}

func TestPromptSpec_Resolve(t *testing.T) {
	base := api.Spec{Model: api.Model{Name: "base-model", Effort: api.EffortMedium}, Budget: api.Budget{Cost: 5}}
	data := map[string]any{"name": "World"}

	t.Run("base model + default body, no override", func(t *testing.T) {
		var op PromptSpec
		got, err := op.Resolve(base, "Summarize {{name}}", data, "")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got.Model.Name != "base-model" {
			t.Errorf("Model = %q, want base-model", got.Model.Name)
		}
		if got.Model.Effort != api.EffortMedium || got.Budget.Cost != 5 {
			t.Errorf("base defaults lost: effort=%q cost=%v", got.Model.Effort, got.Budget.Cost)
		}
		if got.Prompt.User != "Summarize World" {
			t.Errorf("User = %q, want rendered 'Summarize World'", got.Prompt.User)
		}
	})

	t.Run("operation model overrides base, base budget kept", func(t *testing.T) {
		op := PromptSpec{Spec: api.Spec{Model: api.Model{Name: "op-model"}}}
		got, err := op.Resolve(base, "Summarize {{name}}", data, "")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got.Model.Name != "op-model" {
			t.Errorf("Model = %q, want op-model", got.Model.Name)
		}
		if got.Budget.Cost != 5 {
			t.Errorf("Budget.Cost = %v, want 5 from base", got.Budget.Cost)
		}
	})

	t.Run("operation body overrides default body and is rendered", func(t *testing.T) {
		op := PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "Custom {{name}}"}}}
		got, err := op.Resolve(base, "Default {{name}}", data, "")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got.Prompt.User != "Custom World" {
			t.Errorf("User = %q, want 'Custom World'", got.Prompt.User)
		}
	})

	t.Run("built-in default frontmatter model beats base", func(t *testing.T) {
		var op PromptSpec
		def := "---\nmodel: pinned-cheap-model\n---\nSummarize {{name}}"
		got, err := op.Resolve(base, def, data, "")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got.Model.Name != "pinned-cheap-model" {
			t.Errorf("Model = %q, want pinned-cheap-model (default beats base)", got.Model.Name)
		}
	})

	t.Run("operation model beats default frontmatter model", func(t *testing.T) {
		op := PromptSpec{Spec: api.Spec{Model: api.Model{Name: "op-model"}}}
		def := "---\nmodel: pinned-cheap-model\n---\nSummarize {{name}}"
		got, err := op.Resolve(base, def, data, "")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if got.Model.Name != "op-model" {
			t.Errorf("Model = %q, want op-model (op beats default)", got.Model.Name)
		}
	})

	t.Run("no resolvable model fails loud", func(t *testing.T) {
		var op PromptSpec
		_, err := op.Resolve(api.Spec{}, "Summarize {{name}}", data, "")
		if err == nil {
			t.Fatal("expected error for missing model, got nil")
		}
	})
}
