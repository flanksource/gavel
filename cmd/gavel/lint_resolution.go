package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/flanksource/commons/logger"
	deps "github.com/flanksource/deps"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/utils"
)

func resolveLintPath(workDir, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return filepath.Clean(value)
	}

	candidates := make([]string, 0, 2)
	if workDir != "" {
		candidates = append(candidates, filepath.Join(workDir, value))
	}
	if gitRoot := utils.FindGitRoot(workDir); gitRoot != "" && gitRoot != workDir {
		candidates = append(candidates, filepath.Join(gitRoot, value))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return filepath.Clean(value)
}

func normalizeLintRootArg(opts LintOptions) (LintOptions, error) {
	if opts.WorkDir != "" || len(opts.Files) != 1 {
		return opts, nil
	}

	candidate, err := filepath.Abs(opts.Files[0])
	if err != nil {
		return opts, err
	}
	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		return opts, nil
	}

	workDir := candidate
	if goModRoot := utils.FindNearestGoModRoot(candidate); goModRoot != "" {
		workDir = goModRoot
	} else if gitRoot := utils.FindGitRoot(candidate); gitRoot != "" {
		workDir = gitRoot
	}

	opts.WorkDir = workDir
	opts.Files = nil
	return opts, nil
}

func lintGitRoot(workDir string) string {
	if gitRoot := utils.FindGitRoot(workDir); gitRoot != "" {
		return gitRoot
	}
	return workDir
}

func lintToolBinDir(gitRoot string) string {
	return filepath.Join(gitRoot, ".gavel")
}

func executableFileName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) != ".exe" {
		return name + ".exe"
	}
	return name
}

func golangciInstalledPath(gitRoot string) string {
	return filepath.Join(lintToolBinDir(gitRoot), executableFileName("golangci-lint"))
}

func resolveLinterExecutable(ctx context.Context, linter linters.Linter, gitRoot string, hasDirectConfig bool, dryRun bool) (string, string, error) {
	if path, err := exec.LookPath(linter.Name()); err == nil {
		return path, "", nil
	}

	if linter.Name() == "golangci-lint" {
		installed := golangciInstalledPath(gitRoot)
		if info, err := os.Stat(installed); err == nil && !info.IsDir() {
			return installed, "", nil
		}
		if hasDirectConfig {
			if dryRun {
				return installed, "", nil
			}
			binDir := lintToolBinDir(gitRoot)
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				return "", "", fmt.Errorf("create golangci bin dir %s: %w", binDir, err)
			}
			if ctx == nil {
				ctx = context.Background()
			}
			result, err := deps.InstallWithContext(ctx, "golangci/golangci-lint", "stable", deps.WithBinDir(binDir))
			if err != nil {
				return "", "", fmt.Errorf("install golangci-lint in %s: %w", binDir, err)
			}
			if result != nil && result.BinDir != "" {
				installed = filepath.Join(result.BinDir, executableFileName("golangci-lint"))
			}
			if info, err := os.Stat(installed); err == nil && !info.IsDir() {
				logger.V(1).Infof("Resolved golangci-lint to %s", installed)
				return installed, "", nil
			}
			return "", "", fmt.Errorf("golangci-lint install completed but %s was not found", installed)
		}
	}

	return "", "not found on PATH", nil
}
