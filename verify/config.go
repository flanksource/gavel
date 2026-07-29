package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/collections"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/repomap"
	"github.com/ghodss/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

// DefaultAIModel is the built-in global default model used when neither the
// base ai: spec nor an operation spec pins one. It mirrors the value returned
// by ai.DefaultConfig().Model; it is duplicated here as a const so the
// low-level verify package need not import the heavier gavel/ai package.
const DefaultAIModel = "claude-haiku-4-5"

// DefaultAIConfig is the built-in global default base spec (precedence floor).
func DefaultAIConfig() api.Spec {
	return api.Spec{Model: api.Model{Name: DefaultAIModel}}
}

// DefaultGavelConfig seeds the built-in defaults every loaded config layers on.
func DefaultGavelConfig() GavelConfig {
	return GavelConfig{AI: DefaultAIConfig()}
}

// RestartPolicy is the supervisor restart policy for a process. It accepts a
// bool (true→on-failure, false→no) or an enum string (no|on-failure|always)
// from both .gavel.yaml (JSON via ghodss/yaml) and the Procfile (yaml.v3). The
// empty value means "unset" so the resolver applies the default (no).
type RestartPolicy string

const (
	RestartNo        RestartPolicy = "no"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartAlways    RestartPolicy = "always"
)

func (p RestartPolicy) String() string { return string(p) }

// restartPolicyFromString validates an enum string, mapping "" to unset.
func restartPolicyFromString(s string) (RestartPolicy, error) {
	switch RestartPolicy(s) {
	case "", RestartNo, RestartOnFailure, RestartAlways:
		return RestartPolicy(s), nil
	default:
		return "", fmt.Errorf("invalid restart policy %q (want no, on-failure, always, or a bool)", s)
	}
}

func (p *RestartPolicy) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*p = ""
		return nil
	}
	if bytes.Equal(data, []byte("true")) {
		*p = RestartOnFailure
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		*p = RestartNo
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := restartPolicyFromString(s)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

func (p *RestartPolicy) UnmarshalYAML(node *yamlv3.Node) error {
	var b bool
	if err := node.Decode(&b); err == nil {
		if b {
			*p = RestartOnFailure
		} else {
			*p = RestartNo
		}
		return nil
	}
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	v, err := restartPolicyFromString(s)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// LintIgnoreRule suppresses lint violations matching every populated field.
// Rule and Source accept collections.MatchItem patterns: literal strings,
// `*` globs ("acme-*"), and `!`-prefixed negations. File uses doublestar
// globs so directory traversal patterns ("pkg/**/*.go") behave as expected.
// At least one of Rule/Source/File must be set; an empty rule never matches.
type LintIgnoreRule struct {
	Rule   string `yaml:"rule,omitempty" json:"rule,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	File   string `yaml:"file,omitempty" json:"file,omitempty"`
}

func (r LintIgnoreRule) MatchesViolation(v models.Violation) bool {
	if r.Source != "" {
		matched, negated := collections.MatchItem(v.Source, r.Source)
		if negated || !matched {
			return false
		}
	}
	if r.Rule != "" {
		if v.Rule == nil {
			return false
		}
		matched, negated := collections.MatchItem(v.Rule.Method, r.Rule)
		if negated || !matched {
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
	// AIFix is the AI spec used to repair lint violations. It is independent of
	// commit.message because source repair needs an editing-capable agent while
	// commit-message generation only needs to summarize a diff.
	Fix PromptSpec `yaml:"fix,omitempty" json:"fix,omitempty"`
	// Ignore accumulates across config layers: a suppression is a standing
	// statement about this repo, so adding one must not revoke the home config's.
	Ignore  []LintIgnoreRule            `yaml:"ignore,omitempty" json:"ignore,omitempty" merge:"append"`
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

type CheckMode string

func (m CheckMode) String() string {
	return string(m)
}

func (m *CheckMode) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*m = ""
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		*m = "skip"
		return nil
	}
	if bytes.Equal(data, []byte("true")) {
		return fmt.Errorf("expected mode string or false, got true")
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*m = CheckMode(s)
	return nil
}

type CommitConfig struct {
	// Hooks accumulate across config layers, in execution order. GitIgnore and
	// Allow accumulate too and collapse repeats — naming the same path in two
	// configs says nothing more than naming it once.
	Hooks     []CommitHook     `yaml:"hooks,omitempty" json:"hooks,omitempty" merge:"append"`
	GitIgnore []string         `yaml:"gitignore,omitempty" json:"gitignore,omitempty" merge:"append,unique"`
	Allow     []string         `yaml:"allow,omitempty" json:"allow,omitempty" merge:"append,unique"`
	Precommit PrecommitConfig  `yaml:"precommit,omitempty" json:"precommit,omitempty"`
	Lint      CommitLintConfig `yaml:"lint,omitempty" json:"lint,omitempty"`
	Tidy      CommitTidyConfig `yaml:"tidy,omitempty" json:"tidy,omitempty"`
	// Message, Grouping, and Summary are the AI specs (model/prompt/budget/effort)
	// for commit-message generation, AI commit grouping (`gavel commit -G`), and
	// commit-group summaries (`gavel git analyze --summary`). Each overrides the
	// base ai: spec field-wise; empty uses the built-in default.
	Message  PromptSpec `yaml:"message,omitempty" json:"message,omitempty"`
	Grouping PromptSpec `yaml:"grouping,omitempty" json:"grouping,omitempty"`
	Summary  PromptSpec `yaml:"summary,omitempty" json:"summary,omitempty"`
	// MaxCommits caps the number of commits AI grouping may produce (0 = default).
	MaxCommits int `yaml:"maxCommits,omitempty" json:"maxCommits,omitempty"`
}

// CommitTidyConfig controls whether `gavel commit` runs `go mod tidy` in every
// Go module in the repo before committing and stages any go.mod / go.sum
// updates into the in-flight commit. Enabled is on by default (nil = on);
// set to false to disable. CLI flag --tidy overrides per-invocation.
type CommitTidyConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// CommitLintConfig controls whether `gavel commit` runs linters over the
// staged file set before creating the commit. Two independent gates:
//
//   - Enabled toggles every non-secrets linter. nil = off (default), true = on.
//   - Secrets toggles the betterleaks/secrets linter. nil = on (default —
//     secrets are the highest-value pre-commit check), true = on, false = off.
//
// CLI flags --lint and --lint-secrets override these per-invocation.
type CommitLintConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Secrets *bool `yaml:"secrets,omitempty" json:"secrets,omitempty"`
}

// PrecommitConfig configures the combined pre-commit gate for commit.gitignore
// and linked dependency checks. Mode is "prompt" (default), "fail", "skip",
// or false (alias for skip).
type PrecommitConfig struct {
	Mode CheckMode `yaml:"mode,omitempty" json:"mode,omitempty"`
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
	// AI is the base spec (model/fallbacks/budget/effort defaults) that every AI
	// operation inherits and overrides field-wise. See DefaultAIConfig.
	AI       api.Spec       `yaml:"ai,omitempty" json:"ai,omitempty"`
	Lint     LintConfig     `yaml:"lint,omitempty" json:"lint,omitempty"`
	Commit   CommitConfig   `yaml:"commit,omitempty" json:"commit,omitempty"`
	Fixtures FixturesConfig `yaml:"fixtures,omitempty" json:"fixtures,omitempty"`
	SSH      SSHConfig      `yaml:"ssh,omitempty" json:"ssh,omitempty"`
	// Pre and Post accumulate across config layers, in execution order: a repo
	// config adds its steps to the home config's rather than replacing them.
	Pre      []HookStep     `yaml:"pre,omitempty" json:"pre,omitempty" merge:"append"`
	Post     []HookStep     `yaml:"post,omitempty" json:"post,omitempty" merge:"append"`
	Secrets  SecretsConfig  `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Procfile ProcfileConfig `yaml:"procfile,omitempty" json:"procfile,omitempty"`
	// Checks is the project default for the post-completion agent check loop
	// (`gavel todos run --check`). A TODO's frontmatter `checks:` overrides it.
	Checks types.AgentChecksConfig `yaml:"checks,omitempty" json:"checks,omitempty"`
	Todos  TodosConfig             `yaml:"todos,omitempty" json:"todos,omitempty"`
	Status StatusConfig            `yaml:"status,omitempty" json:"status,omitempty"`
	Test   TestConfig              `yaml:"test,omitempty" json:"test,omitempty"`
	PR     PRConfig                `yaml:"pr,omitempty" json:"pr,omitempty"`
}

// PRConfig configures `gavel pr`.
type PRConfig struct {
	// Content is the AI spec for generating the PR title, body, and branch name.
	Content PromptSpec `yaml:"content,omitempty" json:"content,omitempty"`
	// Base is the base branch for the PR (e.g. origin/main).
	Base string `yaml:"base,omitempty" json:"base,omitempty"`
	// Draft opens the PR as a draft.
	Draft bool `yaml:"draft,omitempty" json:"draft,omitempty"`
}

// StatusConfig configures `gavel status`.
type StatusConfig struct {
	// Summary is the AI spec for the per-file summary used by `gavel status --ai`.
	// Empty uses the built-in default. See prompts.StatusSummary.
	Summary PromptSpec `yaml:"summary,omitempty" json:"summary,omitempty"`
}

// TestConfig configures `gavel test`.
type TestConfig struct {
	// OutlineSummary is the AI spec for the per-test summary used by
	// `gavel test outline --ai-summary`. Empty uses the built-in default. See
	// prompts.TestOutlineSummary.
	OutlineSummary PromptSpec `yaml:"outlineSummary,omitempty" json:"outlineSummary,omitempty"`
}

// TodosConfig configures `gavel todos run`.
type TodosConfig struct {
	// Run is the AI spec for the todo run prompt; Plan is the plan-mode spec.
	// Each overrides the base ai: spec field-wise. See prompts.TodosRun/TodosPlan.
	Run  PromptSpec `yaml:"run,omitempty" json:"run,omitempty"`
	Plan PromptSpec `yaml:"plan,omitempty" json:"plan,omitempty"`
	// Driver is the execution mechanism: cmux | cli | sdk | api.
	Driver string `yaml:"driver,omitempty" json:"driver,omitempty"`
	// Timeout caps a run's wall-clock duration (e.g. "30m"); empty = default.
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	// GroupBy selects how todos are grouped into runs.
	GroupBy string `yaml:"groupBy,omitempty" json:"groupBy,omitempty"`
	// Approvals gates Bash behind a human approval prompt. It is a pointer so an
	// explicit `false` turns off an inherited `true`. Only an entrypoint that can
	// answer an approval request may enable it — the CLI has no approver and would
	// block forever, so enabling it there is a loud error rather than a hang.
	// Unset defaults to the entrypoint's own capability (dashboard on, CLI off).
	Approvals *bool `yaml:"approvals,omitempty" json:"approvals,omitempty"`
}

// ProcfileConfig configures `gavel proc` — global defaults for the processes
// declared in the Procfile (per-process settings live in the Procfile itself).
// Every field is optional.
type ProcfileConfig struct {
	// Profile is the default active profile: a Procfile entry with `profiles`
	// auto-starts only when one of them is active. `gavel proc --profile` overrides.
	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`
	// AutoRestart is the default restart policy for every process (bool or enum).
	AutoRestart RestartPolicy `yaml:"autoRestart,omitempty" json:"autoRestart,omitempty"`
	// MaxRestarts caps automatic restarts per process (0 = unlimited).
	MaxRestarts int `yaml:"maxRestarts,omitempty" json:"maxRestarts,omitempty"`
	// Env is injected into every process on top of the parent environment and
	// any sibling .env file.
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	// Mem caps resident memory per process (e.g. "512Mi", "2g"). Empty disables
	// the limit. A process whose tree exceeds it is killed.
	Mem string `yaml:"mem,omitempty" json:"mem,omitempty"`
	// CPU caps sustained CPU per process as a percentage (100 = one full core).
	// 0 disables the limit. A process that stays above it is killed.
	CPU float64 `yaml:"cpu,omitempty" json:"cpu,omitempty"`
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
	// .gavel.yaml's directory). Layers accumulate and repeats collapse.
	Configs []string `yaml:"configs,omitempty" json:"configs,omitempty" merge:"append,unique"`
}

type GavelConfigSource struct {
	Origin string      `json:"origin" yaml:"origin"`
	Path   string      `json:"path" yaml:"path"`
	Raw    string      `json:"-" yaml:"-"`
	Config GavelConfig `json:"config" yaml:"config"`
}

type GavelConfigTrace struct {
	TargetPath string              `json:"targetPath" yaml:"targetPath"`
	TargetDir  string              `json:"targetDir" yaml:"targetDir"`
	GitRoot    string              `json:"gitRoot,omitempty" yaml:"gitRoot,omitempty"`
	Sources    []GavelConfigSource `json:"sources,omitempty" yaml:"sources,omitempty"`
	Merged     GavelConfig         `json:"merged" yaml:"merged"`
}

func LoadGavelConfig(cwd string) (GavelConfig, error) {
	cfg := DefaultGavelConfig()

	home, err := os.UserHomeDir()
	if err == nil {
		if cfg, err = mergeFromFile(cfg, filepath.Join(home, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	gitRoot := repomap.FindGitRoot(cwd)
	if gitRoot != "" {
		if cfg, err = mergeFromFile(cfg, filepath.Join(gitRoot, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	absCwd, _ := filepath.Abs(cwd)
	if absCwd != gitRoot {
		if cfg, err = mergeFromFile(cfg, filepath.Join(absCwd, ".gavel.yaml")); err != nil {
			return GavelConfig{}, err
		}
	}

	return cfg, nil
}

// LoadGavelConfigTrace resolves the effective config for the provided file or
// directory path and records which .gavel.yaml files contributed to the merged
// result. Resolution order matches normal loading: built-in defaults, then the
// user's home config, then the git-root config, then the target directory (or
// the parent directory when the target path is a file).
func LoadGavelConfigTrace(path string) (GavelConfigTrace, error) {
	targetPath, targetDir, err := resolveGavelConfigTarget(path)
	if err != nil {
		return GavelConfigTrace{}, err
	}

	trace := GavelConfigTrace{
		TargetPath: targetPath,
		TargetDir:  targetDir,
		Merged:     DefaultGavelConfig(),
	}

	var candidates []GavelConfigSource
	seen := make(map[string]struct{})
	addCandidate := func(origin, candidatePath string) {
		if candidatePath == "" {
			return
		}
		if _, ok := seen[candidatePath]; ok {
			return
		}
		seen[candidatePath] = struct{}{}
		candidates = append(candidates, GavelConfigSource{
			Origin: origin,
			Path:   candidatePath,
		})
	}

	if home, err := os.UserHomeDir(); err == nil {
		addCandidate("user-home", filepath.Join(home, ".gavel.yaml"))
	}

	trace.GitRoot = repomap.FindGitRoot(targetDir)
	if trace.GitRoot != "" {
		addCandidate("git-root", filepath.Join(trace.GitRoot, ".gavel.yaml"))
	}

	origin := "target-directory"
	if targetPath != targetDir {
		origin = "parent-directory"
	}
	addCandidate(origin, filepath.Join(targetDir, ".gavel.yaml"))

	for _, candidate := range candidates {
		cfg, raw, err := loadSingleGavelConfig(candidate.Path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return GavelConfigTrace{}, err
		}

		candidate.Raw = raw
		candidate.Config = cfg
		trace.Sources = append(trace.Sources, candidate)
		trace.Merged = MergeGavelConfig(trace.Merged, cfg)
	}

	return trace, nil
}

// LoadSingleGavelConfig reads one .gavel.yaml file from the given absolute
// path without layering with home/gitRoot/cwd siblings. Returns a zero-value
// config with os.ErrNotExist when the file is missing so callers can detect
// "need to create" vs. a real read/parse error.
func LoadSingleGavelConfig(path string) (GavelConfig, error) {
	cfg, _, err := loadSingleGavelConfig(path)
	return cfg, err
}

func loadSingleGavelConfig(path string) (GavelConfig, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GavelConfig{}, "", err
	}
	var gc GavelConfig
	if err := yaml.Unmarshal(data, &gc); err != nil {
		return GavelConfig{}, "", fmt.Errorf("parse %s: %w", path, err)
	}
	setPromptSpecBaseDirs(&gc, filepath.Dir(path))
	return gc, string(data), nil
}

func SaveGavelConfig(dir string, cfg GavelConfig) error {
	path := filepath.Join(dir, ".gavel.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// mergeFromFile layers one .gavel.yaml onto base. A missing file is the ordinary
// case — that layer simply does not exist — but every other error is fatal: an
// unreadable or unparseable config, or one carrying an invalid enum, used to be
// discarded here, so a typo anywhere in .gavel.yaml silently ran the whole
// project on built-in defaults instead of the settings it declared.
func mergeFromFile(base GavelConfig, path string) (GavelConfig, error) {
	cfg, err := LoadSingleGavelConfig(path)
	if err != nil {
		if os.IsNotExist(err) {
			return base, nil
		}
		return base, err
	}
	return MergeGavelConfig(base, cfg), nil
}

func resolveGavelConfigTarget(path string) (string, string, error) {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", "", fmt.Errorf("stat %s: %w", absPath, err)
	}

	if info.IsDir() {
		return absPath, absPath, nil
	}

	return absPath, filepath.Dir(absPath), nil
}
