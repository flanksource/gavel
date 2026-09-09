package linters

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func setupGitRepo(root string) {
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func TestFilterViolationsByGitIgnore(t *testing.T) {
	mkViolation := func(file string) models.Violation {
		return models.Violation{File: file, Source: "test"}
	}

	t.Run("removes violations for gitignored files", func(t *testing.T) {
		root := t.TempDir()
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("vendor/\n*.log\n"), 0644)

		violations := []models.Violation{
			mkViolation(filepath.Join(root, "main.go")),
			mkViolation(filepath.Join(root, "vendor", "dep.go")),
			mkViolation(filepath.Join(root, "debug.log")),
		}

		result, err := FilterViolationsByGitIgnore(violations, root)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, filepath.Join(root, "main.go"), result[0].File)
	})

	t.Run("keeps all violations when no git root", func(t *testing.T) {
		root := t.TempDir()
		violations := []models.Violation{
			mkViolation(filepath.Join(root, "a.go")),
			mkViolation(filepath.Join(root, "b.go")),
		}

		result, err := FilterViolationsByGitIgnore(violations, root)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("keeps violations with empty file", func(t *testing.T) {
		root := t.TempDir()
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)

		violations := []models.Violation{
			{Source: "test", Message: models.StringPtr("general warning")},
			mkViolation(filepath.Join(root, "main.go")),
		}

		result, err := FilterViolationsByGitIgnore(violations, root)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("returns empty input unchanged", func(t *testing.T) {
		result, err := FilterViolationsByGitIgnore(nil, t.TempDir())
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("deduplicates file lookups", func(t *testing.T) {
		root := t.TempDir()
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)

		file := filepath.Join(root, "main.go")
		violations := []models.Violation{
			mkViolation(file),
			mkViolation(file),
			mkViolation(filepath.Join(root, "debug.log")),
		}

		result, err := FilterViolationsByGitIgnore(violations, root)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, file, result[0].File)
		assert.Equal(t, file, result[1].File)
	})
}

func TestFilterViolationsByGitIgnoreInResults(t *testing.T) {
	mkViolation := func(file string) models.Violation {
		return models.Violation{File: file, Source: "test"}
	}

	t.Run("filters per-result using each WorkDir", func(t *testing.T) {
		rootA := t.TempDir()
		setupGitRepo(rootA)
		os.WriteFile(filepath.Join(rootA, ".gitignore"), []byte("vendor/\n"), 0644)

		rootB := t.TempDir()
		setupGitRepo(rootB)
		os.WriteFile(filepath.Join(rootB, ".gitignore"), []byte("*.log\n"), 0644)

		results := []*LinterResult{
			{
				Linter:  "a",
				WorkDir: rootA,
				Violations: []models.Violation{
					mkViolation(filepath.Join(rootA, "main.go")),
					mkViolation(filepath.Join(rootA, "vendor", "dep.go")),
				},
			},
			{
				Linter:  "b",
				WorkDir: rootB,
				Violations: []models.Violation{
					mkViolation(filepath.Join(rootB, "app.go")),
					mkViolation(filepath.Join(rootB, "debug.log")),
					mkViolation(filepath.Join(rootB, "trace.log")),
				},
			},
		}

		filtered, err := FilterViolationsByGitIgnoreInResults(results)
		assert.NoError(t, err)
		assert.Equal(t, 3, filtered)
		assert.Len(t, results[0].Violations, 1)
		assert.Equal(t, filepath.Join(rootA, "main.go"), results[0].Violations[0].File)
		assert.Len(t, results[1].Violations, 1)
		assert.Equal(t, filepath.Join(rootB, "app.go"), results[1].Violations[0].File)
	})

	t.Run("returns 0 for nil and empty results", func(t *testing.T) {
		filtered, err := FilterViolationsByGitIgnoreInResults(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, filtered)
		filtered, err = FilterViolationsByGitIgnoreInResults([]*LinterResult{})
		assert.NoError(t, err)
		assert.Equal(t, 0, filtered)
		filtered, err = FilterViolationsByGitIgnoreInResults([]*LinterResult{nil})
		assert.NoError(t, err)
		assert.Equal(t, 0, filtered)
	})

	t.Run("skips results with no violations", func(t *testing.T) {
		root := t.TempDir()
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)

		results := []*LinterResult{
			{Linter: "empty", WorkDir: root, Violations: nil},
		}
		filtered, err := FilterViolationsByGitIgnoreInResults(results)
		assert.NoError(t, err)
		assert.Equal(t, 0, filtered)
	})
}
