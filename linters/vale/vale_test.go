package vale

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The error object vale prints when it finds no configuration.
const noConfigOutput = `{
  "Line": 0,
  "Path": "",
  "Text": "E100 [.vale.ini not found] Runtime error\n\nno config file found\n\nExecution stopped with code 1.",
  "Code": "E100",
  "Span": 0
}`

func TestParseViolations(t *testing.T) {
	t.Run("fails on a vale error object instead of reporting no findings", func(t *testing.T) {
		_, err := NewVale(t.TempDir()).parseViolations([]byte(noConfigOutput))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "E100")
		assert.Contains(t, err.Error(), "no config file found")
	})

	t.Run("returns the alerts of a normal report", func(t *testing.T) {
		output := `{"docs/a.md": [{"Line": 3, "Span": [9, 13], "Check": "Vale.Repetition", "Message": "'is' is repeated!", "Severity": "error"}]}`

		violations, err := NewVale(t.TempDir()).parseViolations([]byte(output))

		require.NoError(t, err)
		require.Len(t, violations, 1)
		assert.Equal(t, 3, violations[0].Line)
		assert.Equal(t, "Vale.Repetition", violations[0].Rule.Method)
	})
}

func TestGeneratedConfig(t *testing.T) {
	t.Setenv("VALE_CONFIG_PATH", "")

	t.Run("writes a config whose styles directory exists when the project has none", func(t *testing.T) {
		workDir := t.TempDir()
		v := NewVale(workDir)

		path, err := v.generatedConfig()

		require.NoError(t, err)
		require.NotEmpty(t, path)
		assert.False(t, strings.HasPrefix(path, workDir), "the config must live outside the work dir")
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(content), "BasedOnStyles = Vale")
		assert.NotContains(t, string(content), "BlockIgnores")
		stylesPath := strings.TrimSpace(strings.TrimPrefix(strings.SplitN(string(content), "\n", 2)[0], "StylesPath ="))
		assert.DirExists(t, stylesPath)
	})

	for _, name := range []string{".vale.ini", "_vale.ini"} {
		t.Run("uses the project's own "+name, func(t *testing.T) {
			workDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(workDir, name), []byte("MinAlertLevel = error\n"), 0o644))
			nested := filepath.Join(workDir, "docs")
			require.NoError(t, os.Mkdir(nested, 0o755))

			path, err := NewVale(nested).generatedConfig()

			require.NoError(t, err)
			assert.Empty(t, path)
		})
	}

	t.Run("uses a --config passed by the user", func(t *testing.T) {
		v := NewVale(t.TempDir())
		v.Config = &models.LinterConfig{Args: []string{"--config=/etc/vale.ini"}}

		path, err := v.generatedConfig()

		require.NoError(t, err)
		assert.Empty(t, path)
	})

	t.Run("uses VALE_CONFIG_PATH", func(t *testing.T) {
		t.Setenv("VALE_CONFIG_PATH", "/etc/vale.ini")

		path, err := NewVale(t.TempDir()).generatedConfig()

		require.NoError(t, err)
		assert.Empty(t, path)
	})
}

func TestBuildArgsExcludesWhenWalking(t *testing.T) {
	t.Run("a directory walk skips the excluded paths at any depth", func(t *testing.T) {
		args := NewVale(t.TempDir()).buildArgsWithConfig("/scratch/.vale.ini")

		require.Equal(t, ".", args[len(args)-1])
		glob := args[len(args)-2]
		assert.True(t, strings.HasPrefix(glob, "--glob=!{"), glob)
		assert.Contains(t, glob, "node_modules/**")
		assert.Contains(t, glob, "**/node_modules/**")
	})

	t.Run("named files are passed as they are", func(t *testing.T) {
		v := NewVale(t.TempDir())
		v.SetOptions(linters.RunOptions{WorkDir: v.WorkDir, Files: []string{"docs/a.md"}})

		assert.Equal(t, []string{"--config=/scratch/.vale.ini", "docs/a.md"}, v.buildArgsWithConfig("/scratch/.vale.ini"))
	})
}
