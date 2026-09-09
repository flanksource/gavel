package lint

import (
	"os"
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/betterleaks"
	"github.com/flanksource/gavel/verify"
)

var linterDirectConfigPatterns = map[string][]string{
	"betterleaks": {
		".betterleaks.toml",
		"betterleaks.toml",
		".gitleaks.toml",
		"gitleaks.toml",
	},
	"eslint": {
		".eslintrc",
		".eslintrc.*",
		"eslint.*",
		"eslint.config.*",
	},
	"golangci-lint": {
		".golangci.yml",
		".golangci.yaml",
		".golangci.toml",
		".golangci.json",
	},
	"markdownlint": {
		".markdownlint.json",
		".markdownlint.jsonc",
		".markdownlint.yaml",
		".markdownlint.yml",
		".markdownlint-cli2.*",
	},
	"oxlint": {
		".oxlintrc.json",
		".oxlintrc.jsonc",
		"oxlint.json",
		"oxlintrc.json",
	},
	"pyright": {
		"pyrightconfig.json",
		"pyproject.toml",
	},
	"react-doctor": {
		"doctor.config.ts",
		"doctor.config.js",
		"doctor.config.mjs",
		"doctor.config.cjs",
		"doctor.config.json",
		"react-doctor.config.json",
	},
	"ruff": {
		"ruff.toml",
		"pyproject.toml",
	},
	"tsc": {
		"tsconfig.json",
	},
	"vale": {
		".vale.ini",
	},
}

func linterConfigPatterns(name string) []string {
	return linterDirectConfigPatterns[name]
}

func linterRequiresDirectConfig(name string) bool {
	return len(linterConfigPatterns(name)) > 0
}

func isLinterExplicitlyEnabled(cfg verify.GavelConfig, name string) bool {
	if cfg.Lint.Linters == nil {
		return false
	}
	linterCfg, ok := cfg.Lint.Linters[name]
	return ok && linterCfg.Enabled != nil && *linterCfg.Enabled
}

func shouldSelectLinter(workDir string, cfg verify.GavelConfig, linter linters.Linter, cliExplicit bool) (bool, string) {
	if cliExplicit {
		if !hasMatchingFiles(workDir, linter.DefaultIncludes(), cfg.Commit.GitIgnore) {
			return false, "no matching files"
		}
		return shouldRunLinter(workDir, cfg, linter.Name(), true, false, false)
	}

	hasConfig := linterHasDirectConfig(workDir, linter)
	if linter.Name() == "betterleaks" {
		hasConfig = len(betterleaks.DiscoverConfigs(workDir)) > 0
	}
	hasDefaultActivation := hasConfig || linterHasDefaultActivation(workDir, linter)
	hasDirectTrigger := hasDirectMatchingFiles(workDir, linter.DefaultIncludes()) || hasDefaultActivation
	if !hasDirectTrigger {
		return false, "no matching files or config in work dir"
	}

	explicitEnabled := isLinterExplicitlyEnabled(cfg, linter.Name())
	return shouldRunLinter(workDir, cfg, linter.Name(), false, explicitEnabled, hasDefaultActivation)
}

func linterHasDirectConfig(workDir string, linter linters.Linter) bool {
	if detector, ok := linter.(linters.DirectConfigDetector); ok {
		return detector.HasDirectConfig(workDir)
	}
	return hasDirectMatchingFiles(workDir, linterConfigPatterns(linter.Name()))
}

func linterHasDefaultActivation(workDir string, linter linters.Linter) bool {
	if detector, ok := linter.(linters.DefaultActivationDetector); ok {
		return detector.HasDefaultActivation(workDir)
	}
	return false
}

func hasDirectMatchingFiles(workDir string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	entries, err := os.ReadDir(workDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		for _, pattern := range patterns {
			if directPatternMatch(name, pattern) {
				return true
			}
		}
	}
	return false
}

func directPatternMatch(name, pattern string) bool {
	if pattern == "" {
		return false
	}
	base := path.Base(strings.ReplaceAll(pattern, "\\", "/"))
	matched, err := doublestar.Match(base, name)
	return err == nil && matched
}
