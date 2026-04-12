package linters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func setupGitRepo(root string) {
	os.MkdirAll(filepath.Join(root, ".git", "info"), 0755)
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

		result := FilterViolationsByGitIgnore(violations, root)
		assert.Len(t, result, 1)
		assert.Equal(t, filepath.Join(root, "main.go"), result[0].File)
	})

	t.Run("keeps all violations when no git root", func(t *testing.T) {
		root := t.TempDir()
		violations := []models.Violation{
			mkViolation(filepath.Join(root, "a.go")),
			mkViolation(filepath.Join(root, "b.go")),
		}

		result := FilterViolationsByGitIgnore(violations, root)
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

		result := FilterViolationsByGitIgnore(violations, root)
		assert.Len(t, result, 2)
	})

	t.Run("returns empty input unchanged", func(t *testing.T) {
		result := FilterViolationsByGitIgnore(nil, "/tmp")
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

		result := FilterViolationsByGitIgnore(violations, root)
		assert.Len(t, result, 2)
		assert.Equal(t, file, result[0].File)
		assert.Equal(t, file, result[1].File)
	})
}
