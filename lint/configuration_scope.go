package lint

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

func partitionInvocations(invocations []linterInvocation, opts Options) ([]linterInvocation, error) {
	var result []linterInvocation
	for _, inv := range invocations {
		configs, err := utils.GlobFilesBounded(inv.projectRoot, []string{"**/.gavel.yaml"})
		if err != nil {
			return nil, err
		}
		var boundaries []string
		for _, config := range configs {
			if dir := filepath.Dir(config); dir != inv.projectRoot {
				boundaries = append(boundaries, dir)
			}
		}
		if len(boundaries) == 0 {
			result = append(result, inv)
			continue
		}
		files, err := utils.GlobFilesBounded(inv.projectRoot, inv.linter.DefaultIncludes())
		if err != nil {
			return nil, err
		}
		buckets := map[string][]string{}
		var order []string
		for _, file := range files {
			if !invocationOwnsFile(inv, invocations, file) {
				continue
			}
			if len(inv.files) > 0 {
				probe := &linters.LinterResult{WorkDir: inv.projectRoot, Violations: nil}
				probe.Violations = append(probe.Violations, models.Violation{File: file})
				linters.FilterViolationsByUserScope([]*linters.LinterResult{probe}, inv.projectRoot, resultScopes(probe, inv.projectRoot, inv.files))
				if len(probe.Violations) == 0 {
					continue
				}
			}
			dir := verify.RuntimeConfigDirectory(inv.projectRoot, opts.ConfigDir)
			for _, boundary := range boundaries {
				if utils.IsWithin(file, boundary) && utils.IsWithin(boundary, dir) {
					dir = boundary
				}
			}
			if _, exists := buckets[dir]; !exists {
				order = append(order, dir)
			}
			rel, err := filepath.Rel(inv.projectRoot, file)
			if err != nil {
				return nil, err
			}
			buckets[dir] = append(buckets[dir], rel)
		}
		for _, dir := range order {
			part := inv
			part.configDir, part.files = dir, buckets[dir]
			if factory, ok := inv.linter.(linters.ProjectFactory); ok {
				part.linter = factory.NewProject(inv.projectRoot)
			}
			result = append(result, part)
		}
	}
	return result, nil
}

func invocationOwnsFile(inv linterInvocation, invocations []linterInvocation, file string) bool {
	for _, other := range invocations {
		if other.linter.Name() == inv.linter.Name() && other.projectRoot != inv.projectRoot && utils.IsWithin(other.projectRoot, inv.projectRoot) && utils.IsWithin(file, other.projectRoot) {

			return false
		}
	}
	return true
}

func (inv linterInvocation) configuration(opts Options) (Options, verify.GavelConfig, linters.RuleExclusions, time.Duration, error) {
	dir := inv.configDir
	if dir == "" {
		dir = verify.RuntimeConfigDirectory(inv.projectRoot, opts.ConfigDir)
	}
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return opts, cfg, nil, 0, err
	}
	opts, err = opts.forLinter(inv.linter.Name(), dir)
	if err != nil {
		return opts, cfg, nil, 0, err
	}
	opts.ConfigDir = dir
	exclusions, err := ParseRuleExclusions(opts.Exclude)
	if err != nil {
		return opts, cfg, nil, 0, err
	}
	timeout, err := parseTimeout(opts.Timeout)
	if timeout == 0 && opts.Timeout != "" {
		disabled := ""
		opts.Runtime.Timeout = &disabled
	}
	return opts, cfg, exclusions, timeout, err
}

func configScopes(inv linterInvocation, scopes []string) []string {
	if inv.configDir == "" {
		return scopes
	}
	result := make([]string, 0, len(inv.files))
	for _, file := range inv.files {
		result = append(result, filepath.Join(inv.projectRoot, file))
	}
	return slices.Clip(result)
}

func filterConfiguredViolations(results []*linters.LinterResult, workDir string) error {
	configs := map[string]verify.GavelConfig{}
	for _, result := range results {
		if result == nil {
			continue
		}
		root := result.WorkDir
		if root == "" {
			root = workDir
		}
		var kept []models.Violation
		for _, violation := range result.Violations {
			dir := root
			if violation.File != "" {
				dir = filepath.Dir(resolveLintPath(root, violation.File))
			}
			cfg, exists := configs[dir]
			if !exists {
				var err error
				cfg, err = verify.LoadGavelConfig(dir)
				if err != nil {
					return err
				}
				configs[dir] = cfg
			}
			probe := &linters.LinterResult{WorkDir: root, Violations: []models.Violation{violation}}
			linters.FilterIgnoredViolations([]*linters.LinterResult{probe}, cfg.Lint.Ignore)
			kept = append(kept, probe.Violations...)
		}
		result.Violations = kept
	}
	return nil
}
