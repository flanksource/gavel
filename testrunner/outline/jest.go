package outline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

func collectJestTests(ctx context.Context, workDir string, filters []string) ([]*Entry, error) {
	packages, err := runners.NewJest(workDir).DiscoverPackages(workDir, true)
	if err != nil {
		return nil, fmt.Errorf("discover jest packages: %w", err)
	}

	var entries []*Entry
	for _, pkg := range packages {
		if !packageMatchesFilters(pkg, filters) {
			continue
		}
		cwd := filepath.Join(workDir, pkg)
		files, err := jestListFiles(ctx, cwd)
		if err != nil {
			entries = append(entries, nodeCollectionError(parsers.Jest, pkg, err))
			continue
		}
		for _, file := range files {
			if !filepath.IsAbs(file) {
				file = filepath.Join(cwd, file)
			}
			rel := relativeTo(file, workDir)
			if !matchesFilters(rel, filters) {
				continue
			}
			source, err := os.ReadFile(file)
			if err != nil {
				entries = append(entries, fileCollectionError(parsers.Jest, rel, err))
				continue
			}
			parsed, err := parseJestSource(rel, source)
			if err != nil {
				entries = append(entries, fileCollectionError(parsers.Jest, rel, err))
				continue
			}
			entries = append(entries, parsed...)
		}
	}
	return entries, nil
}

func jestListFiles(ctx context.Context, cwd string) ([]string, error) {
	command, prefix := runners.DetectPackageManager(cwd)
	args := append(append([]string{}, prefix...), "jest", "--listTests", "--json")
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = cwd
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("jest list in %s failed: %w\nOutput:\n%s", cwd, err, stderr.String())
	}
	var files []string
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("parse jest list output for %s: %w", cwd, err)
	}
	return files, nil
}

func fileCollectionError(framework parsers.Framework, file string, err error) *Entry {
	return &Entry{
		Framework: framework,
		File:      file,
		Name:      "<" + framework.String() + " collection failed>",
		Error:     summarizeCollectionErr(err),
	}
}
