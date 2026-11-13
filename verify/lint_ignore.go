package verify

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

var lintIgnoreMu sync.Mutex

func SaveLintIgnore(directory string, rule LintIgnoreRule) (int, error) {
	if rule.Rule == "" && rule.Source == "" && rule.File == "" && rule.Fingerprint == "" {
		return 0, errors.New("at least one of rule, source, file, or fingerprint is required")
	}
	if rule.Line < 0 || (rule.Line > 0 && rule.File == "") {
		return 0, fmt.Errorf("line %d needs a file and must be positive", rule.Line)
	}
	lintIgnoreMu.Lock()
	defer lintIgnoreMu.Unlock()
	cfg, err := LoadSingleGavelConfig(filepath.Join(directory, GavelConfigFileName))
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	if !slices.ContainsFunc(cfg.Lint.Ignore, func(existing LintIgnoreRule) bool {
		existing.baseDir = ""
		return existing == rule
	}) {
		cfg.Lint.Ignore = append(cfg.Lint.Ignore, rule)
		if err := SaveGavelConfig(directory, cfg); err != nil {
			return 0, err
		}
	}
	return len(cfg.Lint.Ignore), nil
}
