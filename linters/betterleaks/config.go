package betterleaks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
	"github.com/hairyhenderson/toml"
)

// configFilenames is the list of filenames gavel searches for at each
// candidate directory. Both betterleaks and gitleaks ship with their own
// default names; we pick whichever is present.
var configFilenames = []string{
	".betterleaks.toml",
	"betterleaks.toml",
	".gitleaks.toml",
	"gitleaks.toml",
}

// DiscoverConfigs walks the standard gavel config hierarchy
// (home dir → git root → cwd) plus any extra paths listed in the
// `.gavel.yaml` `secrets.configs:` field, and returns every existing
// betterleaks/gitleaks TOML path it finds, in the order gavel should merge
// them (earlier entries are overridden by later ones).
//
// Paths are deduped so the same file referenced twice only appears once.
// Non-existent paths are silently skipped — the caller treats an empty
// result as "no secrets config present, skip the linter".
func DiscoverConfigs(workDir string) []string {
	gc, _ := verify.LoadGavelConfig(workDir)

	var candidates []string
	if home, err := os.UserHomeDir(); err == nil {
		candidates = appendConfigsInDir(candidates, home)
	}
	if gitRoot := repomap.FindGitRoot(workDir); gitRoot != "" {
		candidates = appendConfigsInDir(candidates, gitRoot)
	}
	if abs, err := filepath.Abs(workDir); err == nil {
		candidates = appendConfigsInDir(candidates, abs)
	}

	for _, extra := range gc.Secrets.Configs {
		resolved := extra
		if !filepath.IsAbs(resolved) {
			if gitRoot := repomap.FindGitRoot(workDir); gitRoot != "" {
				resolved = filepath.Join(gitRoot, extra)
			} else {
				resolved = filepath.Join(workDir, extra)
			}
		}
		if _, err := os.Stat(resolved); err == nil {
			candidates = append(candidates, resolved)
		} else {
			logger.V(2).Infof("betterleaks: skipping missing extra config %s", resolved)
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	out := make([]string, 0, len(candidates))
	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}
	return out
}

func appendConfigsInDir(acc []string, dir string) []string {
	for _, name := range configFilenames {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			acc = append(acc, p)
			return acc
		}
	}
	return acc
}

// ResolveConfig returns a path suitable to pass to `betterleaks -c`.
// It takes the list of TOML files discovered at DiscoverConfigs and:
//   - returns "" if no configs exist (caller should skip the linter);
//   - returns the sole config path unchanged when there's only one, so we
//     don't rewrite user files that are already canonical;
//   - otherwise parses each TOML, merges them additively (dedupe rules by
//     ID, union disabledRules, append allowlists, later files win for
//     scalars), writes the result to <workDir>/.tmp/betterleaks.toml, and
//     returns that path.
func ResolveConfig(workDir string, configs []string) (string, error) {
	if len(configs) == 0 {
		return "", nil
	}
	if len(configs) == 1 {
		return configs[0], nil
	}

	var merged tomlConfig
	for _, p := range configs {
		data, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", p, err)
		}
		var cfg tomlConfig
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return "", fmt.Errorf("parse %s: %w", p, err)
		}
		merged = mergeTOMLConfigs(merged, cfg)
	}

	tmpDir := filepath.Join(workDir, ".tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("create .tmp: %w", err)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(merged); err != nil {
		return "", fmt.Errorf("encode merged toml: %w", err)
	}
	out := filepath.Join(tmpDir, "betterleaks.toml")
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", out, err)
	}
	return out, nil
}

// tomlConfig is an INTERNAL minimal mirror of the gitleaks/betterleaks TOML
// schema, used only to merge multiple config files into one. It is NOT a
// user-facing gavel DSL — users author rules in their own .betterleaks.toml
// files and gavel just combines them. Only the fields that need to survive
// round-tripping are modeled; unknown keys are dropped.
type tomlConfig struct {
	Title       string          `toml:"title,omitempty"`
	Description string          `toml:"description,omitempty"`
	Extend      tomlExtend      `toml:"extend,omitempty"`
	Rules       []tomlRule      `toml:"rules,omitempty"`
	Allowlists  []tomlAllowlist `toml:"allowlists,omitempty"`
}

type tomlExtend struct {
	UseDefault    bool     `toml:"useDefault,omitempty"`
	Path          string   `toml:"path,omitempty"`
	URL           string   `toml:"url,omitempty"`
	DisabledRules []string `toml:"disabledRules,omitempty"`
}

type tomlRule struct {
	ID          string          `toml:"id"`
	Description string          `toml:"description,omitempty"`
	Regex       string          `toml:"regex,omitempty"`
	SecretGroup int             `toml:"secretGroup,omitempty"`
	Entropy     float64         `toml:"entropy,omitempty"`
	Path        string          `toml:"path,omitempty"`
	Keywords    []string        `toml:"keywords,omitempty"`
	Tags        []string        `toml:"tags,omitempty"`
	SkipReport  bool            `toml:"skipReport,omitempty"`
	Allowlists  []tomlAllowlist `toml:"allowlists,omitempty"`
}

type tomlAllowlist struct {
	Description string   `toml:"description,omitempty"`
	Condition   string   `toml:"condition,omitempty"`
	RegexTarget string   `toml:"regexTarget,omitempty"`
	TargetRules []string `toml:"targetRules,omitempty"`
	Commits     []string `toml:"commits,omitempty"`
	Paths       []string `toml:"paths,omitempty"`
	Regexes     []string `toml:"regexes,omitempty"`
	StopWords   []string `toml:"stopwords,omitempty"`
}

// mergeTOMLConfigs combines override onto base:
//   - Scalars: later wins when non-zero.
//   - Rules: deduped by ID. Same ID in override replaces the base entry,
//     preserving order for new IDs.
//   - Allowlists: appended (additive — users expect to stack suppression
//     rules across configs, not override them).
//   - Extend.DisabledRules: unioned and deduped.
//   - Extend.UseDefault/Path/URL: last-write-wins (override only if set).
func mergeTOMLConfigs(base, override tomlConfig) tomlConfig {
	if override.Title != "" {
		base.Title = override.Title
	}
	if override.Description != "" {
		base.Description = override.Description
	}

	if override.Extend.UseDefault {
		base.Extend.UseDefault = true
	}
	if override.Extend.Path != "" {
		base.Extend.Path = override.Extend.Path
	}
	if override.Extend.URL != "" {
		base.Extend.URL = override.Extend.URL
	}
	base.Extend.DisabledRules = dedupeStrings(append(base.Extend.DisabledRules, override.Extend.DisabledRules...))

	byID := make(map[string]int, len(base.Rules))
	for i, r := range base.Rules {
		byID[r.ID] = i
	}
	for _, r := range override.Rules {
		if r.ID == "" {
			base.Rules = append(base.Rules, r)
			continue
		}
		if idx, ok := byID[r.ID]; ok {
			base.Rules[idx] = r
		} else {
			byID[r.ID] = len(base.Rules)
			base.Rules = append(base.Rules, r)
		}
	}

	base.Allowlists = append(base.Allowlists, override.Allowlists...)
	return base
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
