package todos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestValidatePlanFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(good, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ValidatePlanFile(good); err != nil {
		t.Errorf("valid plan file rejected: %v", err)
	}
	if err := ValidatePlanFile(""); err == nil {
		t.Error("empty path must be rejected")
	}
	if err := ValidatePlanFile(filepath.Join(dir, "missing.md")); err == nil {
		t.Error("missing file must be rejected")
	}
	if err := ValidatePlanFile(empty); err == nil {
		t.Error("empty file must be rejected")
	}
	if err := ValidatePlanFile(dir); err == nil {
		t.Error("directory must be rejected")
	}
}

func TestHasPlan(t *testing.T) {
	dir := t.TempDir()
	validPlan := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(validPlan, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missingPlan := filepath.Join(dir, "missing.md")

	cases := []struct {
		name string
		todo *types.TODO
		want bool
	}{
		{name: "nil todo", todo: nil, want: false},
		{
			name: "pending, empty plan path",
			todo: &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusPending}},
			want: false,
		},
		{
			name: "pending, missing plan file",
			todo: &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusPending, PlanPath: missingPlan}},
			want: false,
		},
		{
			name: "pending, valid plan file",
			todo: &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusPending, PlanPath: validPlan}},
			want: true,
		},
		{
			name: "review, inline-only plan (no on-disk path)",
			todo: &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusReview}},
			want: true,
		},
		{
			name: "review, invalid plan path still counts (status short-circuits)",
			todo: &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusReview, PlanPath: missingPlan}},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasPlan(tc.todo); got != tc.want {
				t.Errorf("HasPlan() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReadPlanFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(path, []byte("# Plan body"), 0o644); err != nil {
		t.Fatal(err)
	}

	content, _, exists, err := ReadPlanFile(path)
	if err != nil || !exists || content != "# Plan body" {
		t.Errorf("ReadPlanFile = (%q, %t, %v)", content, exists, err)
	}

	// No recorded path and a deleted file both read as "no plan", not an error.
	if _, _, exists, err := ReadPlanFile(""); err != nil || exists {
		t.Errorf("empty path = (%t, %v), want no plan", exists, err)
	}
	if _, _, exists, err := ReadPlanFile(filepath.Join(dir, "gone.md")); err != nil || exists {
		t.Errorf("missing file = (%t, %v), want no plan", exists, err)
	}
}
