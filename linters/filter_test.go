package linters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterIgnoredViolations(t *testing.T) {
	mkViolation := func(source, rule, file string) models.Violation {
		v := models.Violation{Source: source, File: file}
		if rule != "" {
			v.Rule = &models.Rule{Method: rule}
		}
		return v
	}

	tests := []struct {
		name         string
		violations   []models.Violation
		rules        []verify.LintIgnoreRule
		wantKept     int
		wantFiltered int
	}{
		{
			name: "rule only filter",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "a.go"),
				mkViolation("golangci-lint", "unused", "b.go"),
				mkViolation("golangci-lint", "errcheck", "c.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck"}},
			wantKept:     1,
			wantFiltered: 2,
		},
		{
			name: "source only filter",
			violations: []models.Violation{
				mkViolation("eslint", "no-unused-vars", "a.ts"),
				mkViolation("ruff", "F401", "b.py"),
			},
			rules:        []verify.LintIgnoreRule{{Source: "eslint"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name: "rule + file filter",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "pkg/foo.go"),
				mkViolation("golangci-lint", "errcheck", "pkg/bar.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck", File: "pkg/foo.go"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name: "no matches keeps all",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "a.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "unused"}},
			wantKept:     1,
			wantFiltered: 0,
		},
		{
			name: "all filtered",
			violations: []models.Violation{
				mkViolation("eslint", "no-var", "a.ts"),
				mkViolation("eslint", "no-let", "b.ts"),
			},
			rules:        []verify.LintIgnoreRule{{Source: "eslint"}},
			wantKept:     0,
			wantFiltered: 2,
		},
		{
			name:         "empty rules no-op",
			violations:   []models.Violation{mkViolation("ruff", "F401", "a.py")},
			rules:        nil,
			wantKept:     1,
			wantFiltered: 0,
		},
		{
			name: "glob file pattern",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "pkg/sub/deep.go"),
				mkViolation("golangci-lint", "errcheck", "cmd/main.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck", File: "pkg/**/*.go"}},
			wantKept:     1,
			wantFiltered: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LinterResult{Violations: tt.violations}
			filtered := FilterIgnoredViolations([]*LinterResult{result}, tt.rules)
			assert.Equal(t, tt.wantFiltered, filtered)
			assert.Len(t, result.Violations, tt.wantKept)
		})
	}
}

func TestFilterIgnoredViolations_AbsolutePaths(t *testing.T) {
	mkViolation := func(source, rule, file string) models.Violation {
		v := models.Violation{Source: source, File: file}
		if rule != "" {
			v.Rule = &models.Rule{Method: rule}
		}
		return v
	}

	tests := []struct {
		name         string
		workDir      string
		violations   []models.Violation
		rules        []verify.LintIgnoreRule
		wantKept     int
		wantFiltered int
	}{
		{
			name:    "absolute violation path matched by relative ignore rule",
			workDir: "/project",
			violations: []models.Violation{
				mkViolation("markdownlint", "MD034/no-bare-urls", "/project/cmd/hx/fixtures/hx.md"),
				mkViolation("markdownlint", "MD034/no-bare-urls", "/project/docs/other.md"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "MD034/no-bare-urls", Source: "markdownlint", File: "cmd/hx/fixtures/hx.md"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name:    "absolute path with glob pattern",
			workDir: "/project",
			violations: []models.Violation{
				mkViolation("golangci-lint", "errcheck", "/project/pkg/sub/deep.go"),
				mkViolation("golangci-lint", "errcheck", "/project/cmd/main.go"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "errcheck", File: "pkg/**/*.go"}},
			wantKept:     1,
			wantFiltered: 1,
		},
		{
			name:    "relative paths still work",
			workDir: "/project",
			violations: []models.Violation{
				mkViolation("markdownlint", "MD034/no-bare-urls", "cmd/hx/fixtures/hx.md"),
			},
			rules:        []verify.LintIgnoreRule{{Rule: "MD034/no-bare-urls", File: "cmd/hx/fixtures/hx.md"}},
			wantKept:     0,
			wantFiltered: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LinterResult{WorkDir: tt.workDir, Violations: tt.violations}
			filtered := FilterIgnoredViolations([]*LinterResult{result}, tt.rules)
			assert.Equal(t, tt.wantFiltered, filtered)
			assert.Len(t, result.Violations, tt.wantKept)
		})
	}
}

func TestFilterViolationsByUserScope(t *testing.T) {
	// Build a real on-disk project so file-vs-directory detection in the
	// scope filter has something to stat.
	root := t.TempDir()
	mustWrite := func(rel string) string {
		abs := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0o755))
		require.NoError(t, os.WriteFile(abs, nil, 0o644))
		return abs
	}
	targetFile := mustWrite("rules/br/AddressScreen_global.ts")
	siblingFile := mustWrite("rules/br/Other.ts")
	otherDirFile := mustWrite("rules/de/Thing.ts")
	mustWrite("rules/br/nested/Deep.ts")

	mkV := func(file string) models.Violation {
		return models.Violation{Source: "tsc", File: file}
	}

	tests := []struct {
		name        string
		violations  []models.Violation
		scopes      []string
		wantKept    []string
		wantDropped int
	}{
		{
			name: "single file scope keeps only that file",
			violations: []models.Violation{
				mkV(targetFile),
				mkV(siblingFile),
				mkV(otherDirFile),
			},
			scopes:      []string{targetFile},
			wantKept:    []string{targetFile},
			wantDropped: 2,
		},
		{
			name: "directory scope keeps descendants",
			violations: []models.Violation{
				mkV(targetFile),
				mkV(siblingFile),
				mkV(filepath.Join(root, "rules/br/nested/Deep.ts")),
				mkV(otherDirFile),
			},
			scopes: []string{filepath.Join(root, "rules/br")},
			wantKept: []string{
				targetFile,
				siblingFile,
				filepath.Join(root, "rules/br/nested/Deep.ts"),
			},
			wantDropped: 1,
		},
		{
			name: "relative scope resolved against workDir",
			violations: []models.Violation{
				mkV(targetFile),
				mkV(otherDirFile),
			},
			scopes:      []string{"rules/br/AddressScreen_global.ts"},
			wantKept:    []string{targetFile},
			wantDropped: 1,
		},
		{
			name: "relative violation paths resolved against result WorkDir",
			violations: []models.Violation{
				{Source: "tsc", File: "rules/br/AddressScreen_global.ts"},
				{Source: "tsc", File: "rules/de/Thing.ts"},
			},
			scopes:      []string{targetFile},
			wantKept:    []string{"rules/br/AddressScreen_global.ts"},
			wantDropped: 1,
		},
		{
			name:        "empty scopes is no-op",
			violations:  []models.Violation{mkV(targetFile), mkV(otherDirFile)},
			scopes:      nil,
			wantKept:    []string{targetFile, otherDirFile},
			wantDropped: 0,
		},
	}

	t.Run("tsc-style: violations anchored at git root, result.WorkDir is sub-project", func(t *testing.T) {
		// Reproduces the production bug: tsc emits paths relative to the git
		// root, but its result.WorkDir is the tsconfig project root (a
		// subdirectory). The filter must resolve violation paths against the
		// caller's workDir (git root), not just result.WorkDir, or it drops
		// every violation.
		gitRoot := t.TempDir()
		projectRoot := filepath.Join(gitRoot, ".generated")
		require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, "rules/br"), 0o755))
		target := filepath.Join(projectRoot, "rules/br/AddressScreen_global.ts")
		other := filepath.Join(projectRoot, "rules/de/Thing.ts")
		require.NoError(t, os.MkdirAll(filepath.Dir(other), 0o755))
		require.NoError(t, os.WriteFile(target, nil, 0o644))
		require.NoError(t, os.WriteFile(other, nil, 0o644))

		// Violation files are reported relative to gitRoot (as tsc does),
		// while result.WorkDir is projectRoot.
		result := &LinterResult{
			WorkDir: projectRoot,
			Violations: []models.Violation{
				{Source: "tsc", File: ".generated/rules/br/AddressScreen_global.ts"},
				{Source: "tsc", File: ".generated/rules/de/Thing.ts"},
			},
		}
		dropped := FilterViolationsByUserScope(
			[]*LinterResult{result},
			gitRoot,
			[]string{".generated/rules/br/AddressScreen_global.ts"},
		)
		assert.Equal(t, 1, dropped)
		require.Len(t, result.Violations, 1)
		assert.Equal(t, ".generated/rules/br/AddressScreen_global.ts", result.Violations[0].File)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LinterResult{WorkDir: root, Violations: tt.violations}
			dropped := FilterViolationsByUserScope([]*LinterResult{result}, root, tt.scopes)
			assert.Equal(t, tt.wantDropped, dropped)
			gotFiles := make([]string, 0, len(result.Violations))
			for _, v := range result.Violations {
				gotFiles = append(gotFiles, v.File)
			}
			assert.ElementsMatch(t, tt.wantKept, gotFiles)
		})
	}
}
