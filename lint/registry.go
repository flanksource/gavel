package lint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/betterleaks"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/jscpd"
	"github.com/flanksource/gavel/linters/markdownlint"
	"github.com/flanksource/gavel/linters/oxlint"
	"github.com/flanksource/gavel/linters/pyright"
	"github.com/flanksource/gavel/linters/reactdoctor"
	"github.com/flanksource/gavel/linters/ruff"
	"github.com/flanksource/gavel/linters/tsc"
	"github.com/flanksource/gavel/linters/vale"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// buildLinterRegistry registers every available linter rooted at workDir.
// Shared between the execute path and the dry-run path so both stay in sync.
func buildLinterRegistry(workDir string) *linters.Registry {
	registry := linters.NewRegistry()
	registry.Register(golangci.NewGolangciLint(workDir))
	registry.Register(ruff.NewRuff(workDir))
	registry.Register(eslint.NewESLint(workDir))
	registry.Register(oxlint.NewOxlint(workDir))
	registry.Register(reactdoctor.NewReactDoctor(workDir))
	registry.Register(pyright.NewPyright(workDir))
	registry.Register(tsc.NewTSC(workDir))
	registry.Register(markdownlint.NewMarkdownlint(workDir))
	registry.Register(vale.NewVale(workDir))
	registry.Register(jscpd.NewJSCPD(workDir))
	registry.Register(betterleaks.NewBetterleaks(workDir))
	return registry
}

func shouldRunLinter(workDir string, cfg verify.GavelConfig, linterName string, cliExplicit bool, explicitEnabled bool, hasConfig bool) (bool, string) {
	if linterName == "golangci-lint" && utils.FindNearestGoModRoot(workDir) == "" {
		return false, "no go.mod found"
	}
	if linterName == "jscpd" && !cliExplicit && !cfg.Lint.IsLinterEnabled("jscpd", false) {
		return false, "disabled by default; set lint.linters.jscpd.enabled: true"
	}
	if linterName == "betterleaks" {
		if cfg.Secrets.Disabled {
			return false, "disabled via .gavel.yaml"
		}
	}
	if !cliExplicit && !explicitEnabled && linterRequiresDirectConfig(linterName) && !hasConfig {
		if linterName == "betterleaks" {
			return false, "no betterleaks/gitleaks config found"
		}
		if linterName == "react-doctor" {
			return false, "no React dependency or react-doctor config found in work dir"
		}
		return false, fmt.Sprintf("no %s config found in work dir", linterName)
	}
	return true, ""
}

// linterInvocation describes one scheduled linter run against a specific
// project root (go.mod / package.json / tsconfig.json / ...). A linter may
// produce multiple invocations when the input spans several project roots.
type linterInvocation struct {
	linter      linters.Linter
	projectRoot string   // absolute; WorkDir the linter runs from
	files       []string // relative to projectRoot (may be empty = whole root)
}

// resolveLinterInvocations splits one linter across the project roots it
// should run against. When the linter does not implement ProjectRooted, it
// runs once at opts.WorkDir (current behavior). When it does, roots are
// discovered via the input files (or by scanning opts.WorkDir when no files
// were passed) and files are bucketed + relativized per root.
func resolveLinterInvocations(linter linters.Linter, opts Options) []linterInvocation {
	rooted, ok := linter.(linters.ProjectRooted)
	if !ok {
		return []linterInvocation{{linter: linter, projectRoot: opts.WorkDir, files: opts.Files}}
	}
	markers := rooted.ProjectRootMarkers()
	if len(markers) == 0 {
		return []linterInvocation{{linter: linter, projectRoot: opts.WorkDir, files: opts.Files}}
	}

	if len(opts.Files) == 0 {
		roots := utils.FindAllProjectRoots(opts.WorkDir, markers)
		if len(roots) == 0 {
			return nil
		}
		out := make([]linterInvocation, 0, len(roots))
		for _, root := range roots {
			out = append(out, linterInvocation{linter: linter, projectRoot: root})
		}
		return out
	}

	buckets := make(map[string][]string)
	var order []string
	for _, f := range opts.Files {
		abs := resolveLintPath(opts.WorkDir, f)
		dir := abs
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			dir = filepath.Dir(abs)
		}
		root := utils.FindNearestProjectRoot(dir, markers)
		if root == "" {
			logger.V(2).Infof("Skipping %s for %s: no %v found", linter.Name(), f, markers)
			continue
		}
		if _, seen := buckets[root]; !seen {
			order = append(order, root)
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			rel = abs
		}
		buckets[root] = append(buckets[root], rel)
	}
	out := make([]linterInvocation, 0, len(order))
	for _, root := range order {
		out = append(out, linterInvocation{linter: linter, projectRoot: root, files: buckets[root]})
	}
	return out
}

// linterAliases maps CLI-friendly aliases onto registered linter names.
var linterAliases = map[string]string{
	"secrets":       "betterleaks",
	"golangci-lint": "golangci-lint",
	"golangci":      "golangci-lint",
}

// resolveRequestedLinters validates --linters names against the registry and
// returns the canonical list in registry order. Empty input returns every
// registered linter (explicit=false). Any unknown name returns an error with
// the known list so typos fail loudly instead of running the wrong subset.
func resolveRequestedLinters(registry *linters.Registry, requested []string) ([]string, bool, error) {
	if len(requested) == 0 {
		return registry.List(), false, nil
	}
	seen := make(map[string]bool, len(requested))
	out := make([]string, 0, len(requested))
	for _, raw := range requested {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if canonical, ok := linterAliases[name]; ok {
			name = canonical
		}
		if !registry.Has(name) {
			return nil, false, fmt.Errorf("unknown linter %q; known: %s", raw, strings.Join(registry.List(), ", "))
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out, true, nil
}

// hasMatchingFiles checks if any files in workDir match at least one of the glob
// patterns, skipping anything excluded by .gitignore or the .gavel.yaml gitignore
// list (extraIgnore). Gitignored directories are pruned, never descended, so a
// stray match inside vendor/ or node_modules/ never selects a linter. Bails on
// the first match.
func hasMatchingFiles(workDir string, patterns, extraIgnore []string) bool {
	return utils.AnyGlobMatchGitIgnored(workDir, patterns, extraIgnore)
}

// lintBaseRef returns the git ref to use for --new-from-rev computation, or
// "" when neither --since nor --changed was set. --since wins if both are
// present.
func lintBaseRef(opts Options) string {
	if opts.Since != "" {
		return opts.Since
	}
	if opts.Changed {
		if v := os.Getenv("GAVEL_CHANGED_BASE"); v != "" {
			return v
		}
		return "origin/main"
	}
	return ""
}

// resolveMergeBase returns the merge-base commit between HEAD and ref.
// Mirrors golangci-lint's own --new-from-rev semantics: issues present at
// merge-base are ignored, only regressions relative to it are reported.
func resolveMergeBase(workDir, ref string) (string, error) {
	cmd := exec.Command("git", "merge-base", "HEAD", ref)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git merge-base HEAD %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}
