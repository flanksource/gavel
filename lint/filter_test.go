package lint

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/betterleaks"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
)

func newTestLinterRegistry(t *testing.T) *linters.Registry {
	t.Helper()
	r := linters.NewRegistry()
	r.Register(golangci.NewGolangciLint(""))
	r.Register(eslint.NewESLint(""))
	r.Register(betterleaks.NewBetterleaks(""))
	return r
}

func TestCheckPassThrough(t *testing.T) {
	tests := []struct {
		name        string
		requested   []string
		passThrough []string
		wantErr     string
	}{
		{name: "no args with many linters", requested: []string{"eslint", "golangci-lint"}},
		{name: "args with one linter", requested: []string{"golangci-lint"}, passThrough: []string{"--enable", "gosec"}},
		{name: "args with many linters", requested: []string{"eslint", "golangci-lint"}, passThrough: []string{"--fix"}, wantErr: "require exactly one linter (got eslint, golangci-lint)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPassThrough(tt.requested, tt.passThrough)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestResolveRequestedLinters(t *testing.T) {
	r := newTestLinterRegistry(t)

	t.Run("empty returns all, not explicit", func(t *testing.T) {
		got, explicit, err := resolveRequestedLinters(r, nil)
		if err != nil {
			t.Fatal(err)
		}
		if explicit {
			t.Error("expected explicit=false for empty input")
		}
		if len(got) != len(r.List()) {
			t.Errorf("got %d, want %d", len(got), len(r.List()))
		}
	})

	t.Run("single name is explicit", func(t *testing.T) {
		got, explicit, err := resolveRequestedLinters(r, []string{"eslint"})
		if err != nil {
			t.Fatal(err)
		}
		if !explicit {
			t.Error("expected explicit=true")
		}
		if len(got) != 1 || got[0] != "eslint" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("secrets alias resolves to betterleaks", func(t *testing.T) {
		got, _, err := resolveRequestedLinters(r, []string{"secrets"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0] != "betterleaks" {
			t.Errorf("got %v, want [betterleaks]", got)
		}
	})

	t.Run("golangci alias resolves to golangci-lint", func(t *testing.T) {
		got, _, err := resolveRequestedLinters(r, []string{"golangci"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0] != "golangci-lint" {
			t.Errorf("got %v, want [golangci-lint]", got)
		}
	})

	t.Run("dedupes duplicates", func(t *testing.T) {
		got, _, err := resolveRequestedLinters(r, []string{"eslint", "eslint", "secrets", "betterleaks"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("got %v, want 2 entries", got)
		}
	})

	t.Run("unknown hard-fails with known list", func(t *testing.T) {
		_, _, err := resolveRequestedLinters(r, []string{"nope"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "unknown linter") {
			t.Errorf("error should say unknown linter: %q", err)
		}
		for _, name := range r.List() {
			if !strings.Contains(err.Error(), name) {
				t.Errorf("error should list %q: %q", name, err)
			}
		}
	})
}
