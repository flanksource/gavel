package vale

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/linters"
)

// valeConfigNames are the project config files vale itself looks for in the
// directory it runs from and every parent.
var valeConfigNames = []string{".vale.ini", "_vale.ini"}

// GenerateValeConfig writes the .vale.ini gavel uses for a project that has
// none: Vale's built-in style (spelling, repetition, terms) for prose files.
// It lives in gavel's scratch dir, outside workDir, with an empty styles
// directory beside it because vale refuses a StylesPath that does not exist.
func GenerateValeConfig(workDir string) (string, error) {
	dir, err := linters.ScratchDir(workDir, "vale")
	if err != nil {
		return "", err
	}
	stylesPath := filepath.Join(dir, "styles")
	if err := os.MkdirAll(stylesPath, 0o755); err != nil {
		return "", fmt.Errorf("create Vale styles dir: %w", err)
	}

	content := fmt.Sprintf(`StylesPath = %s
MinAlertLevel = suggestion

[*.{md,mdx,markdown,txt,rst,adoc}]
BasedOnStyles = Vale
`, stylesPath)

	configPath := filepath.Join(dir, ".vale.ini")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write Vale config: %w", err)
	}
	return configPath, nil
}

// hasProjectValeConfig reports whether vale would find a config of its own for
// a run in workDir: VALE_CONFIG_PATH, or a config file in workDir or a parent.
func hasProjectValeConfig(workDir string) (bool, error) {
	if os.Getenv("VALE_CONFIG_PATH") != "" {
		return true, nil
	}
	dir, err := filepath.Abs(workDir)
	if err != nil {
		return false, fmt.Errorf("resolve Vale work dir %q: %w", workDir, err)
	}
	for {
		for _, name := range valeConfigNames {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true, nil
			} else if !os.IsNotExist(err) {
				return false, fmt.Errorf("check Vale config %s: %w", filepath.Join(dir, name), err)
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false, nil
		}
		dir = parent
	}
}

// walkExcludeGlob is vale's --glob for a directory walk: every exclude as one
// negated alternation, each matched at the root and at any depth.
func walkExcludeGlob(excludes []string) string {
	patterns := make([]string, 0, len(excludes)*2)
	for _, exclude := range excludes {
		exclude = strings.TrimPrefix(exclude, "/")
		patterns = append(patterns, exclude)
		if !strings.HasPrefix(exclude, "**/") {
			patterns = append(patterns, "**/"+exclude)
		}
	}
	return "--glob=!{" + strings.Join(patterns, ",") + "}"
}
