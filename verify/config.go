package verify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/repomap"
	"github.com/ghodss/yaml"
)

type ChecksConfig struct {
	Disabled           []string `yaml:"disabled" json:"disabled"`
	DisabledCategories []string `yaml:"disabledCategories" json:"disabledCategories"`
}

type VerifyConfig struct {
	Model  string       `yaml:"model" json:"model"`
	Prompt string       `yaml:"prompt" json:"prompt"`
	Checks ChecksConfig `yaml:"checks" json:"checks"`
}

type LintIgnoreRule struct {
	Rule   string `yaml:"rule,omitempty" json:"rule,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	File   string `yaml:"file,omitempty" json:"file,omitempty"`
}

func (r LintIgnoreRule) MatchesViolation(v models.Violation) bool {
	if r.Source != "" && r.Source != v.Source {
		return false
	}
	if r.Rule != "" {
		if v.Rule == nil || v.Rule.Method != r.Rule {
			return false
		}
	}
	if r.File != "" {
		matched, _ := doublestar.Match(r.File, v.File)
		if !matched {
			return false
		}
	}
	return r.Rule != "" || r.Source != "" || r.File != ""
}

type LintConfig struct {
	Ignore  []LintIgnoreRule            `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	Linters map[string]LintLinterConfig `yaml:"linters,omitempty" json:"linters,omitempty"`
}

type LintLinterConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

func (c LintConfig) IsLinterEnabled(name string, defaultEnabled bool) bool {
	if c.Linters == nil {
		return defaultEnabled
	}
	cfg, ok := c.Linters[name]
	if !ok || cfg.Enabled == nil {
		return defaultEnabled
	}
	return *cfg.Enabled
}

type CommitHook struct {
	Name  string   `yaml:"name" json:"name"`
	Run   string   `yaml:"run" json:"run"`
	Files []string `yaml:"files,omitempty" json:"files,omitempty"`
}

type CommitConfig struct {
	Model      string           `yaml:"model,omitempty" json:"model,omitempty"`
	Hooks      []CommitHook     `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	GitIgnore  []string         `yaml:"gitignore,omitempty" json:"gitignore,omitempty"`
	Allow      []string         `yaml:"allow,omitempty" json:"allow,omitempty"`
	LinkedDeps LinkedDepsConfig `yaml:"linkedDeps,omitempty" json:"linkedDeps,omitempty"`
}

// LinkedDepsConfig configures the pre-commit check that blocks go.mod
// replace directives and package.json file:/link: references pointing
// outside the git root. Mode is "prompt" (default), "fail", or "skip".
type LinkedDepsConfig struct {
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`
}

// HookStep is a single shell command rendered into the SSH post-receive hook.
// Used by top-level Pre/Post in GavelConfig.
type HookStep struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Run  string `yaml:"run" json:"run"`
}

// SSHConfig overrides the main command run by the SSH post-receive hook.
// When Cmd is empty, the hook falls back to `gavel test --lint`.
type SSHConfig struct {
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
}

// DefaultFixturesGlob is the default glob pattern used to discover fixture files.
const DefaultFixturesGlob = "**/*.fixture.md"

type FixturesConfig struct {
	Enabled bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Files   []string `yaml:"files,omitempty" json:"files,omitempty"`
}

// ResolvedFiles returns the configured globs, falling back to the default when none are set.
func (f FixturesConfig) ResolvedFiles() []string {
	if len(f.Files) > 0 {
		return f.Files
	}
	return []string{DefaultFixturesGlob}
}

type GavelConfig struct {
	Verify   VerifyConfig   `yaml:"verify" json:"verify"`
	Lint     LintConfig     `yaml:"lint,omitempty" json:"lint,omitempty"`
	Commit   CommitConfig   `yaml:"commit,omitempty" json:"commit,omitempty"`
	Fixtures FixturesConfig `yaml:"fixtures,omitempty" json:"fixtures,omitempty"`
	SSH      SSHConfig      `yaml:"ssh,omitempty" json:"ssh,omitempty"`
	Pre      []HookStep     `yaml:"pre,omitempty" json:"pre,omitempty"`
	Post     []HookStep     `yaml:"post,omitempty" json:"post,omitempty"`
	Secrets  SecretsConfig  `yaml:"secrets,omitempty" json:"secrets,omitempty"`
}

// SecretsConfig turns the betterleaks linter on/off and optionally points at
// extra betterleaks/gitleaks TOML configs beyond the ones gavel discovers
// from the home dir, git root, and cwd. Rule authoring lives in those TOML
// files, not here — gavel only orchestrates discovery + merge.
type SecretsConfig struct {
	// Disabled turns off the betterleaks linter even when the binary is on
	// PATH. Defaults to false (enabled).
	Disabled bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	// Configs is an optional list of additional .betterleaks.toml /
	// .gitleaks.toml paths to merge in (relative paths resolve against the
	// .gavel.yaml's directory).
	Configs []string `yaml:"configs,omitempty" json:"configs,omitempty"`
}

func DefaultVerifyConfig() VerifyConfig {
	return VerifyConfig{
		Model: "claude",
	}
}

func LoadConfig(cwd string) (VerifyConfig, error) {
	gc, err := LoadGavelConfig(cwd)
	return gc.Verify, err
}

func LoadGavelConfig(cwd string) (GavelConfig, error) {
	cfg := GavelConfig{Verify: DefaultVerifyConfig()}

	home, err := os.UserHomeDir()
	if err == nil {
		cfg = mergeFromFile(cfg, filepath.Join(home, ".gavel.yaml"))
	}

	gitRoot := repomap.FindGitRoot(cwd)
	if gitRoot != "" {
		cfg = mergeFromFile(cfg, filepath.Join(gitRoot, ".gavel.yaml"))
	}

	absCwd, _ := filepath.Abs(cwd)
	if absCwd != gitRoot {
		cfg = mergeFromFile(cfg, filepath.Join(absCwd, ".gavel.yaml"))
	}

	return cfg, nil
}

// LoadSingleGavelConfig reads one .gavel.yaml file from the given absolute
// path without layering with home/gitRoot/cwd siblings. Returns a zero-value
// config with os.ErrNotExist when the file is missing so callers can detect
// "need to create" vs. a real read/parse error.
func LoadSingleGavelConfig(path string) (GavelConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GavelConfig{}, err
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return GavelConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return gc, nil
}

func SaveGavelConfig(dir string, cfg GavelConfig) error {
	path := filepath.Join(dir, ".gavel.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func mergeFromFile(base GavelConfig, path string) GavelConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return base
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return base
	}
	base.Verify = MergeVerifyConfig(base.Verify, gc.Verify)
	base.Lint = MergeLintConfig(base.Lint, gc.Lint)
	base.Commit = MergeCommitConfig(base.Commit, gc.Commit)
	base.Fixtures = MergeFixturesConfig(base.Fixtures, gc.Fixtures)
	base.SSH = MergeSSHConfig(base.SSH, gc.SSH)
	base.Pre = append(base.Pre, gc.Pre...)
	base.Post = append(base.Post, gc.Post...)
	base.Secrets = MergeSecretsConfig(base.Secrets, gc.Secrets)
	return base
}

// MergeSecretsConfig merges override onto base. Disabled is OR (any layer
// disabling wins). Configs are appended and deduped so each TOML path only
// appears once even when multiple .gavel.yaml files reference it.
func MergeSecretsConfig(base, override SecretsConfig) SecretsConfig {
	if override.Disabled {
		base.Disabled = true
	}
	seen := make(map[string]struct{}, len(base.Configs)+len(override.Configs))
	var merged []string
	for _, p := range append(base.Configs, override.Configs...) {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		merged = append(merged, p)
	}
	base.Configs = merged
	return base
}

// MergeSSHConfig merges override onto base. Cmd is last-write-wins; an empty
// override preserves the base value so a repo config can inherit the home
// default.
func MergeSSHConfig(base, override SSHConfig) SSHConfig {
	if override.Cmd != "" {
		base.Cmd = override.Cmd
	}
	return base
}

func MergeVerifyConfig(base, override VerifyConfig) VerifyConfig {
	if override.Model != "" {
		base.Model = override.Model
	}
	if override.Prompt != "" {
		base.Prompt = override.Prompt
	}
	if len(override.Checks.Disabled) > 0 {
		base.Checks.Disabled = append(base.Checks.Disabled, override.Checks.Disabled...)
	}
	if len(override.Checks.DisabledCategories) > 0 {
		base.Checks.DisabledCategories = append(base.Checks.DisabledCategories, override.Checks.DisabledCategories...)
	}
	return base
}

func MergeLintConfig(base, override LintConfig) LintConfig {
	if len(override.Ignore) > 0 {
		base.Ignore = append(base.Ignore, override.Ignore...)
	}
	if len(override.Linters) > 0 {
		if base.Linters == nil {
			base.Linters = make(map[string]LintLinterConfig, len(override.Linters))
		}
		for name, cfg := range override.Linters {
			merged := base.Linters[name]
			if cfg.Enabled != nil {
				merged.Enabled = cfg.Enabled
			}
			base.Linters[name] = merged
		}
	}
	return base
}

func MergeCommitConfig(base, override CommitConfig) CommitConfig {
	if override.Model != "" {
		base.Model = override.Model
	}
	if len(override.Hooks) > 0 {
		base.Hooks = append(base.Hooks, override.Hooks...)
	}
	base.GitIgnore = dedupStrings(append(base.GitIgnore, override.GitIgnore...))
	base.Allow = dedupStrings(append(base.Allow, override.Allow...))
	if override.LinkedDeps.Mode != "" {
		base.LinkedDeps.Mode = override.LinkedDeps.Mode
	}
	return base
}

func dedupStrings(in []string) []string {
	if len(in) == 0 {
		return nil
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

// MergeFixturesConfig merges override onto base. Enabled is true if either side
// sets it; Files from the override replace base so a repo-level config can
// override a home-level default without accumulating globs.
func MergeFixturesConfig(base, override FixturesConfig) FixturesConfig {
	if override.Enabled {
		base.Enabled = true
	}
	if len(override.Files) > 0 {
		base.Files = override.Files
	}
	return base
}
