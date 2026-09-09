package outline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

type playwrightListEnvelope struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func collectPlaywrightTests(ctx context.Context, workDir string, filters []string) ([]*Entry, error) {
	packages, err := runners.NewPlaywright(workDir).DiscoverPackages(workDir, true)
	if err != nil {
		return nil, fmt.Errorf("discover playwright packages: %w", err)
	}

	var entries []*Entry
	for _, pkg := range packages {
		if !packageMatchesFilters(pkg, filters) {
			continue
		}
		cwd := filepath.Join(workDir, pkg)
		data, runErr := playwrightList(ctx, cwd)
		listed, reportedErrors, parseErr := parsePlaywrightList(data, cwd, pkg, workDir)
		if parseErr != nil {
			if runErr != nil {
				entries = append(entries, nodeCollectionError(parsers.Playwright, pkg, runErr))
			} else {
				entries = append(entries, nodeCollectionError(parsers.Playwright, pkg, parseErr))
			}
			continue
		}
		for _, entry := range listed {
			if matchesFilters(entry.File, filters) {
				entries = append(entries, entry)
			}
		}
		for _, message := range reportedErrors {
			entries = append(entries, nodeCollectionError(parsers.Playwright, pkg, fmt.Errorf("%s", message)))
		}
		if runErr != nil && len(reportedErrors) == 0 {
			entries = append(entries, nodeCollectionError(parsers.Playwright, pkg, runErr))
		}
	}
	return entries, nil
}

func playwrightList(ctx context.Context, cwd string) ([]byte, error) {
	command, prefix := runners.DetectPackageManager(cwd)
	args := append(append([]string{}, prefix...), "playwright", "test", "--list", "--reporter=json")
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = cwd
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return output, fmt.Errorf("playwright list in %s failed: %w\nOutput:\n%s", cwd, err, stderr.String())
	}
	return output, nil
}

func parsePlaywrightList(data []byte, cwd, pkg, workDir string) ([]*Entry, []string, error) {
	var envelope playwrightListEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, nil, fmt.Errorf("parse playwright list output for %s: %w", cwd, err)
	}

	tests, err := parsers.NewPlaywrightJSON(cwd).Parse(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	entries := make([]*Entry, 0, len(tests))
	for _, test := range tests {
		test.Framework = parsers.Playwright
		entries = append(entries, entryFromParsedTest(test, pkg, workDir))
	}
	errors := make([]string, 0, len(envelope.Errors))
	for _, reported := range envelope.Errors {
		message := strings.TrimSpace(parsers.StripANSI(reported.Message))
		if message != "" {
			errors = append(errors, message)
		}
	}
	return entries, errors, nil
}
