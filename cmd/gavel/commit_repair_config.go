package main

import (
	"fmt"
	"strconv"

	"github.com/flanksource/gavel/verify"
)

func commitRepairConfig(opts CommitOptions, cfg verify.CommitLintConfig) (verify.CommitLintConfig, error) {
	if opts.AIFix != "" {
		value, err := strconv.ParseBool(opts.AIFix)
		if err != nil {
			return cfg, fmt.Errorf("--ai-fix requires true or false: %w", err)
		}
		cfg.AIFix = &value
	}
	if opts.AIFixMaxIterations != 0 {
		cfg.AIFixMaxIterations = &opts.AIFixMaxIterations
	}
	if cfg.AIFixMaxIterations != nil && *cfg.AIFixMaxIterations < 1 {
		return cfg, fmt.Errorf("commit.lint.aiFixMaxIterations must be positive")
	}
	return cfg, nil
}
