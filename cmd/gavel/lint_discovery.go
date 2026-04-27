package main

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
	"pyright": {
		"pyrightconfig.json",
		"pyproject.toml",
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
		if !hasMatchingFiles(workDir, linter.DefaultIncludes()) {
			return false, "no matching files"
		}
		return shouldRunLinter(workDir, cfg, linter.Name(), true, false, false)
	}

	hasConfig := hasDirectMatchingFiles(workDir, linterConfigPatterns(linter.Name()))
	if linter.Name() == "betterleaks" {
		hasConfig = len(betterleaks.DiscoverConfigs(workDir)) > 0
	}
	hasDirectTrigger := hasDirectMatchingFiles(workDir, linter.DefaultIncludes()) || hasConfig
	if !hasDirectTrigger {
		return false, "no matching files or config in work dir"
	}

	explicitEnabled := isLinterExplicitlyEnabled(cfg, linter.Name())
	return shouldRunLinter(workDir, cfg, linter.Name(), false, explicitEnabled, hasConfig)
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
